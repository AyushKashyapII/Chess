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

    // Helper function to convert row/col to algebraic notation (like engine_cli.go)
    function coordsToSquare(row, col) {
        const file = String.fromCharCode('a'.charCodeAt(0) + col);
        const rank = 8 - row;
        return file + rank;
    }

    async function validateMove(move) {
        try {
            // Convert coordinates to string format like "e2e4" (same as engine_cli.go)
            const fromSquareStr = coordsToSquare(move.FromRow, move.FromCol);
            const toSquareStr = coordsToSquare(move.ToRow, move.ToCol);
            const moveString = fromSquareStr + toSquareStr;
            
            console.log('Sending move string:', moveString);
            
            const result = await callWorker("VALIDATE_MOVE", {
                moveString: moveString
            });

            console.log('Validation result (Worker):', result);

            if (result && result.valid) {
                if (result.newFen) {
                    boardState = fenToBoard(result.newFen);
                    console.log('Board updated after user move, new FEN:', result.newFen);
                } else {
                    console.error('No newFen in validation result!');
                }
                fromSquare = null;
                updateUi();

                setTimeout(() => {
                    getAiMove();
                }, 100);
            } else {
                statusElement.textContent = result.error || 'Invalid Move! Try again.';
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

    // Helper function to split array 
    function splitIntoChunks(array, numChunks) {
        const chunks = [];
        const chunkSize = Math.ceil(array.length / numChunks);
        for (let i = 0; i < array.length; i += chunkSize) {
            chunks.push(array.slice(i, i + chunkSize));
        }
        return chunks;
    }

    // Parallel search 
    async function getAiMove() {
        isAwaitingAi = true;
        updateStatus();

        try {
            const fen = boardToFen();
            console.log('Starting AI move search with FEN:', fen);
            const numWorkers = window.chessWorkers.length;

            await Promise.all(window.chessWorkers.map((worker, index) => {
                return new Promise((resolve) => {
                    const listener = (e) => {
                        if (e.data.type === "INIT_BOARD_RESULT") {
                            worker.removeEventListener("message", listener);
                            resolve();
                        }
                    };
                    worker.addEventListener("message", listener);
                    worker.postMessage({ 
                        type: "INIT_BOARD", 
                        payload: { fen: fen } 
                    });
                });
            }));
            
            console.log(`All ${numWorkers} workers initialized with FEN:`, fen);
            
            const allMovesJson = await new Promise((resolve) => {
                const listener = (e) => {
                    if (e.data.type === "GET_MOVES_RESULT") {
                        window.chessWorkers[0].removeEventListener("message", listener);
                        resolve(e.data.data);
                    }
                };
                window.chessWorkers[0].addEventListener("message", listener);
                window.chessWorkers[0].postMessage({ 
                    type: "GET_ALL_MOVES", 
                    fen: fen 
                });
            });

            const allMoves = JSON.parse(allMovesJson);
            console.log(`Found ${allMoves.length} legal moves, splitting across ${numWorkers} workers`);

            if (allMoves.length === 0) {
                statusElement.textContent = 'Game Over! You Won';
                isGameOver = true;
                isAwaitingAi = false;
                updateUi();
                return;
            }

            // Split moves into chunks
            const chunks = splitIntoChunks(allMoves, numWorkers);
            
            // Send chunks to workers
            let workerResults = [];
            let completedWorkers = 0;

            chunks.forEach((chunk, i) => {
                if (chunk.length === 0) return; // Skip empty chunks
                
                const worker = window.chessWorkers[i];
                const listener = (e) => {
                    if (e.data.type === "SEARCH_SUBSET_RESULT") {
                        worker.removeEventListener("message", listener);
                        workerResults.push(e.data.data);
                        completedWorkers++;
                        
                        // All workers finished
                        if (completedWorkers === chunks.filter(c => c.length > 0).length) {
                            const bestOverall = workerResults.reduce((prev, current) => 
                                (current.score < prev.score) ? current : prev
                            );
                            
                            console.log('Best move from parallel search:', bestOverall);
                            
                            // Apply the move using WASM (handles castling, en passant, promotion)
                            if (bestOverall.move && bestOverall.fromRow !== undefined) {
                                const moveJson = JSON.stringify({
                                    fromRow: bestOverall.fromRow,
                                    fromCol: bestOverall.fromCol,
                                    toRow: bestOverall.toRow,
                                    toCol: bestOverall.toCol
                                });
                                
                                // Apply move through worker to get proper new FEN
                                window.chessWorkers[0].postMessage({ 
                                    type: "APPLY_MOVE", 
                                    fen: fen,
                                    moveJson: moveJson
                                });
                                
                                const applyMoveListener = (e) => {
                                    if (e.data.type === "APPLY_MOVE_RESULT") {
                                        window.chessWorkers[0].removeEventListener("message", applyMoveListener);
                                        
                                        if (e.data.data.newFen) {
                                            // Update board state from new FEN
                                            const newFen = e.data.data.newFen;
                                            boardState = fenToBoard(newFen);
                                            console.log('Board updated after AI move, new FEN:', newFen);
                                            
                                            // Sync all workers with new board state (async, don't wait)
                                            window.chessWorkers.forEach(worker => {
                                                worker.postMessage({ 
                                                    type: "INIT_BOARD", 
                                                    payload: { fen: newFen } 
                                                });
                                            });
                                            
                                            // Check if white has moves left (white's turn = true)
                                            window.chessWorkers[0].postMessage({ 
                                                type: "GET_ALL_MOVES", 
                                                fen: newFen,
                                                isWhiteTurn: true
                                            });
                                            
                                            const whiteMovesListener = (e2) => {
                                                if (e2.data.type === "GET_MOVES_RESULT") {
                                                    window.chessWorkers[0].removeEventListener("message", whiteMovesListener);
                                                    const whiteMoves = JSON.parse(e2.data.data);
                                                    
                                                    if (whiteMoves.length === 0) {
                                                        statusElement.textContent = "Game Over! You Lost";
                                                        isGameOver = true;
                                                    }
                                                    
                                                    console.log('Engine plays:', bestOverall.move);
                                                    isAwaitingAi = false;
                                                    updateUi();
                                                }
                                            };
                                            window.chessWorkers[0].addEventListener("message", whiteMovesListener);
                                        } else {
                                            console.error('Error applying move:', e.data.data);
                                            // Fallback: use single worker
                                            getAiMoveSingleWorker();
                                        }
                                    }
                                };
                                window.chessWorkers[0].addEventListener("message", applyMoveListener);
                            } else {
                                // Fallback: use single worker
                                getAiMoveSingleWorker();
                            }
                        }
                    }
                };
                worker.addEventListener("message", listener);
                worker.postMessage({ 
                    type: "SEARCH_SUBSET", 
                    fen: fen, 
                    movesToSearch: chunk 
                });
            });

        } catch (error) {
            console.error('Error in parallel search, falling back to single worker:', error);
            getAiMoveSingleWorker();
        }
    }

    // Helper to convert square to row/col
    function squareToRow(square) {
        return 8 - parseInt(square[1]);
    }

    function squareToCol(square) {
        return square.charCodeAt(0) - 'a'.charCodeAt(0);
    }

    // Fallback: single worker search (original method)
    async function getAiMoveSingleWorker() {
        try {
            // Initialize worker with current FEN before getting AI move
            const fen = boardToFen();
            await new Promise((resolve) => {
                const listener = (e) => {
                    if (e.data.type === "INIT_BOARD_RESULT") {
                        window.chessWorker.removeEventListener("message", listener);
                        resolve();
                    }
                };
                window.chessWorker.addEventListener("message", listener);
                window.chessWorker.postMessage({ 
                    type: "INIT_BOARD", 
                    payload: { fen: fen } 
                });
            });
            
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
                if (aiMove.move) {
                    console.log('Engine plays:', aiMove.move);
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