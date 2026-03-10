document.addEventListener('DOMContentLoaded', () => {
    const boardElement = document.getElementById('board');
    const statusElement = document.getElementById('status');
    const restartButton = document.getElementById('restart-button');
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

    // Helper to send messages to the worker and return a promise
    function callWorker(type, payload) {
        console.log(`Main: Sending ${type}`, payload);
        return new Promise((resolve) => {
            const listener = (e) => {
                if (e.data.type === type + "_RESULT") {
                    window.chessWorker.removeEventListener("message", listener);
                    resolve(e.data.payload);
                }
            };
            window.chessWorker.addEventListener("message", listener);
            window.chessWorker.postMessage({ type, payload });
        });
    }

    function handleSquareClick(row, col) {
        console.log('Clicked:', row, col);

        if (isAwaitingAi) {
            console.log('Waiting for AI, click ignored');
            return;
        }

        const clickedPiece = boardState[row][col];
        const isWhitePiece = clickedPiece !== ' ' && clickedPiece === clickedPiece.toUpperCase();
        if (fromSquare === null) {
            if (isWhitePiece) {
                fromSquare = { row, col };
                console.log('Selected piece at:', row, col);
                updateUi();
            }
            return;
        }
        if (fromSquare.row === row && fromSquare.col === col) {
            console.log('Deselecting piece');
            fromSquare = null;
            updateUi();
            return;
        }

        const selectedPiece = boardState[fromSquare.row][fromSquare.col];
        const targetPiece = boardState[row][col];
        const isTargetWhite = targetPiece !== ' ' && targetPiece === targetPiece.toUpperCase();
        if (targetPiece !== ' ' && isWhitePiece && isTargetWhite) {
            fromSquare = { row, col };
            console.log('Re-selected different piece at:', row, col);
            updateUi();
            return;
        }
        const move = {
            FromRow: fromSquare.row,
            FromCol: fromSquare.col,
            ToRow: row,
            ToCol: col
        };

        console.log('Attempting move:', move);
        isAwaitingAi = true;
        statusElement.textContent = 'Validating move...';
        validateMove(move);
    }

    async function validateMove(move) {
        try {
            const result = await callWorker("VALIDATE_MOVE", {
                fromR: move.FromRow,
                fromC: move.FromCol,
                toR: move.ToRow,
                toC: move.ToCol
            });

            console.log('Validation result (Worker):', result);

            if (result && result.valid) {
                if (result.newFen) {
                    boardState = fenToBoard(result.newFen);
                } else {
                    // This branch should ideally not be hit if worker always returns newFen
                    // but kept for robustness if worker logic changes.
                    // makeMove(move); // Removed as worker handles board state
                }
                fromSquare = null;
                updateUi();

                setTimeout(() => {
                    getAiMove();
                }, 100);
            } else {
                statusElement.textContent = 'Invalid Move! Try again.';
                isAwaitingAi = false;
                updateUi();
            }

        } catch (error) {
            console.error('Error during worker communication:', error);
            statusElement.textContent = 'Error: ' + error.message;
            isAwaitingAi = false;
            fromSquare = null;
            updateUi();
        }
    }

    async function getAiMove() {
        isAwaitingAi = true;
        updateStatus();

        try {
            const aiMove = await callWorker("GET_AI_MOVE", {});
            console.log('AI move received (Worker):', aiMove);

            if (aiMove && aiMove.valid) {
                if (!aiMove.gamestatus) {
                    statusElement.textContent = "Game Over! You Lost";
                    isGameOver = true;
                }
                if (aiMove.newFen) {
                    boardState = fenToBoard(aiMove.newFen);
                }
            } else {
                statusElement.textContent = 'Game Over! You Won';
                isGameOver = true;
            }

        } catch (error) {
            console.error('Error getting AI move:', error);
            statusElement.textContent = 'Error: ' + error.message;
        } finally {
            isAwaitingAi = false;
            updateUi();
        }
    }

    function updateUi() {
        console.log('Updating UI...');
        renderBoard();
        updateStatus();
    }

    function renderBoard() {
        console.log('Rendering board...');
        boardElement.innerHTML = '';

        for (let r = 0; r < 8; r++) {
            for (let c = 0; c < 8; c++) {
                const square = document.createElement('div');
                square.classList.add('square', (r + c) % 2 === 0 ? 'light' : 'dark');
                square.dataset.row = r;
                square.dataset.col = c;

                if (fromSquare && fromSquare.row === r && fromSquare.col === c) {
                    square.classList.add('selected');
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
        console.log('Board rendered');
    }

    function updateStatus() {
        if (isGameOver) {
            statusElement.textContent = "Game Over!!!"
        } else if (isAwaitingAi) {
            statusElement.textContent = 'Black is thinking...';
        } else {
            statusElement.textContent = 'White to move';
        }
    }

    // makeMove is no longer needed as the worker handles board state updates via FEN
    // function makeMove(move) {
    //     console.log('makeMove called with:', move);
    //     console.log('Board before:', JSON.stringify(boardState));

    //     const piece = boardState[move.FromRow][move.FromCol];
    //     console.log('Moving piece:', piece, 'from', move.FromRow, move.FromCol, 'to', move.ToRow, move.ToCol);

    //     boardState[move.ToRow][move.ToCol] = piece;
    //     boardState[move.FromRow][move.FromCol] = ' ';

    //     console.log('Board after:', JSON.stringify(boardState));
    // }

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
            if (empty > 0) {
                fen += empty;
            }
            if (r < 7) {
                fen += '/';
            }
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

    function initGame() {
        console.log('Initializing game...');
        const startFen = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR';
        boardState = fenToBoard(startFen);

        // Sync the worker state
        if (window.chessWorker) {
            window.chessWorker.postMessage({ type: "INIT_BOARD", payload: { fen: startFen } });
        }

        fromSquare = null;
        isAwaitingAi = false;
        isGameOver = false;
        statusElement.textContent = 'White to move';
        updateUi();
    }

    // Wait for worker to be ready before starting game first time
    window.onChessWorkerReady = () => {
        initGame();
    };

    restartButton.addEventListener('click', initGame);
    // Don't call initGame() directly here anymore, wasm-init.js will call window.onChessWorkerReady
});