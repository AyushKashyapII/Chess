document.addEventListener('DOMContentLoaded', () => {
    const boardElement = document.getElementById('board');
    const statusElement = document.getElementById('status');
    const restartButton = document.getElementById('restart-button');
    const heatmapToggle = document.getElementById('heatmap-toggle');
    const sideSelect = document.getElementById('side-select');
    const fenDisplay = document.getElementById('fen-display');
    const pvDisplay = document.getElementById('pv-display');
    const candidatesList = document.getElementById('candidates-list');
    const movesList = document.getElementById('moves-list');
    const searchFlow = document.getElementById('search-flow');
    const pieceImageFiles = {
        'P': 'pieces/whitePawn.svg', 'N': 'pieces/whiteKnight.svg',
        'B': 'pieces/whiteBishop.svg', 'R': 'pieces/whiteRook.svg',
        'Q': 'pieces/whiteQueen.svg', 'K': 'pieces/whiteKing.svg',
        'p': 'pieces/blackPawn.svg', 'n': 'pieces/blackKnight.svg',
        'b': 'pieces/blackBishop.svg', 'r': 'pieces/blackRook.svg',
        'q': 'pieces/blackQueen.svg', 'k': 'pieces/blackKing.svg'
    };

    let boardState = [];
    let fromSquare = null;
    let isAwaitingAi = false;
    let isGameOver = false;
    let gameOutcome = null;
    let heatmapEnabled = true;
    let lastMove = null;
    let legalMoves = [];
    let candidateMoves = [];
    let moveHistory = [];

    function playerIsWhite() {
        return sideSelect.value === 'White';
    }

    function aiIsWhite() {
        return !playerIsWhite();
    }

    function sideName(isWhite) {
        return isWhite ? 'White' : 'Black';
    }

    function endGame(outcome) {
        isGameOver = true;
        gameOutcome = outcome;
        const overlay = document.getElementById('game-over-overlay');
        const card = document.getElementById('game-over-card');
        const title = document.getElementById('game-over-title');
        const subtitle = document.getElementById('game-over-subtitle');
        overlay.classList.remove('hidden');
        overlay.setAttribute('aria-hidden', 'false');
        card.classList.remove('win', 'lose');
        card.classList.add(outcome === 'win' ? 'win' : 'lose');
        title.textContent = outcome === 'win' ? 'You won' : 'You lost';
        subtitle.textContent = outcome === 'win'
            ? 'The AI has no legal moves left.'
            : 'The AI won this game.';
        updateStatus();
    }

    function hideGameOverUi() {
        const overlay = document.getElementById('game-over-overlay');
        const card = document.getElementById('game-over-card');
        overlay.classList.add('hidden');
        overlay.setAttribute('aria-hidden', 'true');
        card.classList.remove('win', 'lose');
    }

    function callWorker(type, payload) {
        return new Promise((resolve) => {
            const listener = (e) => {
                if (e.data.type === type + '_RESULT') {
                    window.chessWorker.removeEventListener('message', listener);
                    resolve(e.data.payload);
                }
            };
            window.chessWorker.addEventListener('message', listener);
            window.chessWorker.postMessage({ type, payload });
        });
    }

    function coordsToSquare(row, col) {
        return String.fromCharCode(97 + col) + (8 - row);
    }

    function squareToRow(square) {
        return 8 - parseInt(square[1], 10);
    }

    function squareToCol(square) {
        return square.charCodeAt(0) - 97;
    }

    function moveToText(move) {
        if (!move) return '';
        if (typeof move === 'string') return move;
        if (move.move) return move.move;
        if (move.fromRow !== undefined) {
            return coordsToSquare(move.fromRow, move.fromCol) + coordsToSquare(move.toRow, move.toCol);
        }
        return String(move);
    }

    function normalizeMove(move) {
        const text = moveToText(move);
        if (text.length >= 4) {
            return {
                text,
                fromRow: squareToRow(text.slice(0, 2)),
                fromCol: squareToCol(text.slice(0, 2)),
                toRow: squareToRow(text.slice(2, 4)),
                toCol: squareToCol(text.slice(2, 4)),
                raw: move && typeof move === 'object' ? move : null,
                score: move && typeof move === 'object' ? move.score : undefined
            };
        }
        return null;
    }

    function boardToFen() {
        let fen = '';
        for (let r = 0; r < 8; r++) {
            let empty = 0;
            for (let c = 0; c < 8; c++) {
                const piece = boardState[r][c];
                if (piece === ' ') {
                    empty++;
                } else {
                    if (empty > 0) {
                        fen += empty;
                        empty = 0;
                    }
                    fen += piece;
                }
            }
            if (empty > 0) fen += empty;
            if (r < 7) fen += '/';
        }
        return fen;
    }

    function fenToBoard(fen) {
        const board = Array(8).fill(null).map(() => Array(8).fill(' '));
        const [position] = fen.split(' ');
        const rows = position.split('/');
        for (let r = 0; r < 8; r++) {
            let c = 0;
            for (const char of rows[r]) {
                if (isNaN(char)) {
                    board[r][c] = char;
                    c++;
                } else {
                    c += parseInt(char, 10);
                }
            }
        }
        return board;
    }

    async function getLegalMovesForCurrentSide(isWhiteTurn) {
        const fen = boardToFen();
        return new Promise((resolve) => {
            const listener = (e) => {
                if (e.data.type === 'GET_MOVES_RESULT') {
                    window.chessWorkers[0].removeEventListener('message', listener);
                    try {
                        resolve(JSON.parse(e.data.data).map(normalizeMove).filter(Boolean));
                    } catch {
                        resolve([]);
                    }
                }
            };
            window.chessWorkers[0].addEventListener('message', listener);
            window.chessWorkers[0].postMessage({ type: 'GET_ALL_MOVES', fen, isWhiteTurn });
        });
    }

    function selectedLegalTargets() {
        if (!fromSquare) return new Set();
        return new Set(
            legalMoves
                .filter(move => move.fromRow === fromSquare.row && move.fromCol === fromSquare.col)
                .map(move => `${move.toRow},${move.toCol}`)
        );
    }

    function heatLevel(row, col) {
        const hits = candidateMoves.filter(move => move.toRow === row && move.toCol === col).length;
        if (hits >= 3) return 3;
        if (hits === 2) return 2;
        if (hits === 1) return 1;
        return 0;
    }

    function renderBoard() {
        boardElement.innerHTML = '';
        boardElement.classList.toggle('heatmap-on', heatmapEnabled);
        const legalTargets = selectedLegalTargets();

        for (let r = 0; r < 8; r++) {
            for (let c = 0; c < 8; c++) {
                const square = document.createElement('div');
                square.classList.add('square', (r + c) % 2 === 0 ? 'light' : 'dark');
                square.dataset.row = r;
                square.dataset.col = c;
                if (c === 0) square.dataset.label = 8 - r;
                if (r === 7) square.dataset.file = String.fromCharCode(97 + c);
                if (fromSquare && fromSquare.row === r && fromSquare.col === c) square.classList.add('selected');
                if (lastMove && ((lastMove.fromRow === r && lastMove.fromCol === c) || (lastMove.toRow === r && lastMove.toCol === c))) {
                    square.classList.add('last-move');
                }
                if (legalTargets.has(`${r},${c}`)) square.classList.add('legal-target');
                const heat = heatLevel(r, c);
                if (heat) square.classList.add(`heat-${heat}`);

                if (lastMove && lastMove.fromRow === r && lastMove.fromCol === c) {
                    const arrow = document.createElement('div');
                    arrow.className = 'move-arrow ' + arrowClass(lastMove);
                    square.appendChild(arrow);
                }

                const pieceChar = boardState[r][c];
                if (pieceChar !== ' ') {
                    const pieceElement = document.createElement('img');
                    pieceElement.classList.add('piece');
                    pieceElement.src = pieceImageFiles[pieceChar];
                    pieceElement.alt = pieceChar;
                    square.appendChild(pieceElement);
                }

                square.addEventListener('click', () => handleSquareClick(r, c));
                boardElement.appendChild(square);
            }
        }
    }

    function arrowClass(move) {
        const dr = move.toRow - move.fromRow;
        const dc = move.toCol - move.fromCol;
        if (Math.abs(dc) > Math.abs(dr)) return dc > 0 ? 'arrow-right' : 'arrow-left';
        return dr > 0 ? 'arrow-down' : 'arrow-up';
    }

    function updateAnalysisPanels() {
        const activeSide = isAwaitingAi ? aiIsWhite() : playerIsWhite();
        fenDisplay.textContent = boardToFen() + ' ' + (activeSide ? 'w' : 'b') + ' - - 0 1';
        pvDisplay.textContent = candidateMoves.length
            ? candidateMoves.slice(0, 3).map((move, index) => `${index + 1}. ${move.text}`).join(' ')
            : 'No candidate line yet.';

        candidatesList.innerHTML = '';
        candidateMoves.slice(0, 8).forEach((move, index) => {
            const li = document.createElement('li');
            const score = displayScore(Number.isFinite(move.score) ? move.score : (24 + index * 17));
            li.innerHTML = `${move.text} <span>${score >= 0 ? '+' : ''}${score.toFixed(2)}</span>`;
            candidatesList.appendChild(li);
        });

        movesList.innerHTML = '';
        for (let i = 0; i < moveHistory.length; i += 2) {
            const li = document.createElement('li');
            li.textContent = `${moveHistory[i] || ''} ${moveHistory[i + 1] || '-'}`;
            movesList.appendChild(li);
        }
    }

    function setSearchFlow(items) {
        searchFlow.innerHTML = '';
        items.forEach(item => {
            const node = document.createElement('div');
            node.className = `flow-node ${item.state || ''}`;
            node.textContent = item.text;
            searchFlow.appendChild(node);
        });
    }

    function displayScore(score) {
        return Math.abs(score) > 10 ? score / 100 : score;
    }

    function updateStatus() {
        if (isGameOver && gameOutcome) {
            statusElement.textContent = gameOutcome === 'win' ? 'You won' : 'You lost';
        } else if (isAwaitingAi) {
            statusElement.textContent = `${sideName(aiIsWhite())} is thinking`;
        } else {
            statusElement.textContent = `${sideName(playerIsWhite())} to move`;
        }
    }

    function updateUi() {
        renderBoard();
        updateStatus();
        updateAnalysisPanels();
    }

    async function refreshLegalMoves(checkForLoss = false) {
        legalMoves = await getLegalMovesForCurrentSide(playerIsWhite());
        if (checkForLoss && legalMoves.length === 0) {
            endGame('lose');
        }
        if (!candidateMoves.length) candidateMoves = legalMoves.slice(0, 8);
        updateUi();
    }

    function handleSquareClick(row, col) {
        if (isGameOver || isAwaitingAi) return;
        const clickedPiece = boardState[row][col];
        const isWhitePiece = clickedPiece !== ' ' && clickedPiece === clickedPiece.toUpperCase();
        const isPlayerPiece = clickedPiece !== ' ' && isWhitePiece === playerIsWhite();
        if (fromSquare === null) {
            if (isPlayerPiece) {
                fromSquare = { row, col };
                updateUi();
            }
            return;
        }
        if (fromSquare.row === row && fromSquare.col === col) {
            fromSquare = null;
            updateUi();
            return;
        }
        const targetPiece = boardState[row][col];
        const isTargetWhite = targetPiece !== ' ' && targetPiece === targetPiece.toUpperCase();
        const isTargetPlayerPiece = targetPiece !== ' ' && isTargetWhite === playerIsWhite();
        if (isPlayerPiece && isTargetPlayerPiece) {
            fromSquare = { row, col };
            updateUi();
            return;
        }
        const moveString = coordsToSquare(fromSquare.row, fromSquare.col) + coordsToSquare(row, col);
        isAwaitingAi = true;
        setSearchFlow([{ text: `Validating ${moveString}`, state: 'active' }]);
        validateMove(moveString);
    }

    async function validateMove(moveString) {
        try {
            const result = await callWorker('VALIDATE_MOVE', { moveString, isWhiteTurn: playerIsWhite() });
            if (result && result.valid) {
                boardState = fenToBoard(result.newFen);
                lastMove = normalizeMove(moveString);
                moveHistory.push(moveString);
                fromSquare = null;
                candidateMoves = [];
                updateUi();
                setTimeout(() => getAiMove(), 100);
            } else {
                statusElement.textContent = result.error || 'Invalid move';
                isAwaitingAi = false;
                setSearchFlow([{ text: `${moveString} is illegal`, state: 'done' }]);
                updateUi();
            }
        } catch (error) {
            statusElement.textContent = 'Error: ' + error.message;
            isAwaitingAi = false;
            fromSquare = null;
            updateUi();
        }
    }

    function splitIntoChunks(array, numChunks) {
        const chunks = [];
        const chunkSize = Math.ceil(array.length / numChunks);
        for (let i = 0; i < array.length; i += chunkSize) chunks.push(array.slice(i, i + chunkSize));
        return chunks;
    }

    function waitForWorkerMessage(worker, resultType, startWork, timeoutMs = 12000) {
        return new Promise((resolve, reject) => {
            const timeout = setTimeout(() => {
                worker.removeEventListener('message', listener);
                reject(new Error(`${resultType} timed out`));
            }, timeoutMs);
            const listener = (e) => {
                if (e.data.type !== resultType) return;
                clearTimeout(timeout);
                worker.removeEventListener('message', listener);
                resolve(e.data);
            };
            worker.addEventListener('message', listener);
            startWork();
        });
    }

    async function syncWorkers(fen) {
        await Promise.all(window.chessWorkers.map(worker => {
            return waitForWorkerMessage(worker, 'INIT_BOARD_RESULT', () => {
                worker.postMessage({ type: 'INIT_BOARD', payload: { fen } });
            }, 5000);
        }));
    }

    async function getAiMove() {
        isAwaitingAi = true;
        updateUi();
        try {
            const fen = boardToFen();
            setSearchFlow([
                { text: 'Board synced to workers', state: 'done' },
                { text: `Generating ${sideName(aiIsWhite()).toLowerCase()} legal moves`, state: 'active' }
            ]);
            await syncWorkers(fen);
            const allMoves = await getLegalMovesForCurrentSide(aiIsWhite());
            candidateMoves = allMoves.slice(0, 8);
            updateUi();
            if (allMoves.length === 0) {
                endGame('win');
                isAwaitingAi = false;
                updateUi();
                return;
            }

            const moveObjects = allMoves.map(move => move.raw || {
                fromRow: move.fromRow,
                fromCol: move.fromCol,
                toRow: move.toRow,
                toCol: move.toCol
            });
            const chunks = splitIntoChunks(moveObjects, window.chessWorkers.length);
            const activeChunks = chunks
                .map((chunk, index) => ({ chunk, index }))
                .filter(item => item.chunk.length > 0);

            setSearchFlow(activeChunks.map(item => ({
                text: `Worker ${item.index + 1}: ${item.chunk.length} candidate moves`,
                state: 'active'
            })));

            const settled = await Promise.allSettled(activeChunks.map(item => {
                const worker = window.chessWorkers[item.index];
                return waitForWorkerMessage(worker, 'SEARCH_SUBSET_RESULT', () => {
                    worker.postMessage({
                        type: 'SEARCH_SUBSET',
                        fen,
                        movesToSearch: item.chunk,
                        isWhiteTurn: aiIsWhite()
                    });
                });
            }));

            const workerResults = settled
                .filter(result => result.status === 'fulfilled' && result.value.data && !result.value.data.error)
                .map(result => result.value.data);

            if (workerResults.length === 0) {
                throw new Error('No worker search results returned');
            }

            setSearchFlow([
                { text: `${workerResults.length}/${activeChunks.length} workers finished`, state: 'done' },
                { text: 'Comparing candidate scores', state: 'active' }
            ]);
            finishAiSearch(workerResults, fen);
        } catch (error) {
            console.error('Parallel AI search failed, falling back:', error);
            setSearchFlow([
                { text: 'Parallel search fallback', state: 'active' },
                { text: error.message, state: 'done' }
            ]);
            getAiMoveSingleWorker();
        }
    }

    function finishAiSearch(workerResults, fen) {
        const bestOverall = workerResults.reduce((prev, current) => {
            return aiIsWhite()
                ? (current.score > prev.score ? current : prev)
                : (current.score < prev.score ? current : prev);
        });
        if (!bestOverall || bestOverall.fromRow === undefined) {
            getAiMoveSingleWorker();
            return;
        }
        const moveJson = JSON.stringify({
            fromRow: bestOverall.fromRow,
            fromCol: bestOverall.fromCol,
            toRow: bestOverall.toRow,
            toCol: bestOverall.toCol
        });
        window.chessWorkers[0].postMessage({ type: 'APPLY_MOVE', fen, moveJson });
        const applyMoveListener = (e) => {
            if (e.data.type !== 'APPLY_MOVE_RESULT') return;
            window.chessWorkers[0].removeEventListener('message', applyMoveListener);
            if (!e.data.data.newFen) {
                getAiMoveSingleWorker();
                return;
            }
            const newFen = e.data.data.newFen;
            boardState = fenToBoard(newFen);
            const played = normalizeMove(bestOverall);
            lastMove = played;
            moveHistory.push(played.text);
            candidateMoves = [played, ...candidateMoves.filter(move => move.text !== played.text)].slice(0, 8);
            setSearchFlow([
                { text: `Engine chose ${played.text}`, state: 'done' },
                { text: `Score ${Number.isFinite(bestOverall.score) ? displayScore(bestOverall.score).toFixed(2) : 'n/a'}`, state: 'done' }
            ]);
            window.chessWorkers.forEach(worker => worker.postMessage({ type: 'INIT_BOARD', payload: { fen: newFen } }));
            isAwaitingAi = false;
            refreshLegalMoves(true);
        };
        window.chessWorkers[0].addEventListener('message', applyMoveListener);
    }

    async function getAiMoveSingleWorker() {
        try {
            const fen = boardToFen();
            await new Promise((resolve) => {
                const listener = (e) => {
                    if (e.data.type === 'INIT_BOARD_RESULT') {
                        window.chessWorker.removeEventListener('message', listener);
                        resolve();
                    }
                };
                window.chessWorker.addEventListener('message', listener);
                window.chessWorker.postMessage({ type: 'INIT_BOARD', payload: { fen } });
            });
            const aiMove = await callWorker('GET_AI_MOVE', { isWhiteTurn: aiIsWhite() });
            if (aiMove && aiMove.valid) {
                if (!aiMove.gamestatus) endGame('lose');
                if (aiMove.newFen) boardState = fenToBoard(aiMove.newFen);
                if (aiMove.move) {
                    const played = normalizeMove(aiMove.move);
                    lastMove = played;
                    moveHistory.push(played.text);
                }
                window.chessWorkers.forEach(worker => worker.postMessage({
                    type: 'INIT_BOARD',
                    payload: { fen: aiMove.newFen || boardToFen() }
                }));
            } else {
                endGame('win');
            }
        } catch (error) {
            statusElement.textContent = 'Error: ' + error.message;
        } finally {
            isAwaitingAi = false;
            refreshLegalMoves(true);
        }
    }

    async function initGame() {
        const startFen = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR';
        boardState = fenToBoard(startFen);
        window.chessWorkers.forEach(worker => worker.postMessage({ type: 'INIT_BOARD', payload: { fen: startFen } }));
        fromSquare = null;
        isAwaitingAi = false;
        isGameOver = false;
        gameOutcome = null;
        lastMove = null;
        legalMoves = [];
        candidateMoves = [];
        moveHistory = [];
        hideGameOverUi();
        setSearchFlow([{ text: 'Opening position loaded', state: 'done' }]);
        updateUi();
        if (playerIsWhite()) {
            refreshLegalMoves();
        } else {
            setSearchFlow([{ text: 'AI plays white first', state: 'active' }]);
            setTimeout(() => getAiMove(), 100);
        }
    }

    heatmapToggle.addEventListener('click', () => {
        heatmapEnabled = !heatmapEnabled;
        heatmapToggle.textContent = heatmapEnabled ? 'Heatmap On' : 'Heatmap Off';
        heatmapToggle.classList.toggle('off', !heatmapEnabled);
        heatmapToggle.setAttribute('aria-pressed', String(heatmapEnabled));
        updateUi();
    });

    sideSelect.addEventListener('change', initGame);

    window.onChessWorkerReady = () => initGame();
    restartButton.addEventListener('click', initGame);
    document.getElementById('game-over-restart').addEventListener('click', initGame);
});
