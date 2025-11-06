document.addEventListener('DOMContentLoaded', () => {
    // --- DOM Elements ---
    const boardElement = document.getElementById('board');
    const statusElement = document.getElementById('status');
    const restartButton = document.getElementById('restart-button');

    // --- Piece Image Mapping ---
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

    function handleSquareClick(row, col) {
        console.log('Clicked:', row, col);
        
        if (isAwaitingAi) {
            console.log('Waiting for AI, click ignored');
            return;
        }

        const clickedPiece = boardState[row][col];
        const isWhitePiece = clickedPiece !== ' ' && clickedPiece === clickedPiece.toUpperCase();

        // No piece selected yet
        if (fromSquare === null) {
            if (isWhitePiece) {
                fromSquare = { row, col };
                console.log('Selected piece at:', row, col);
                updateUi();
            }
            return;
        }

        // Double-click on same square to deselect
        if (fromSquare.row === row && fromSquare.col === col) {
            console.log('Deselecting piece');
            fromSquare = null;
            updateUi();
            return;
        }

        const selectedPiece = boardState[fromSquare.row][fromSquare.col];
        const targetPiece = boardState[row][col];
        const isTargetWhite = targetPiece !== ' ' && targetPiece === targetPiece.toUpperCase();

        // Can't capture your own piece - allow re-selection instead
        if (targetPiece !== ' ' && isWhitePiece && isTargetWhite) {
            fromSquare = { row, col };
            console.log('Re-selected different piece at:', row, col);
            updateUi();
            return;
        }

        // Attempt the move
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
        const fenBefore = boardToFen();
        console.log('FEN before move:', fenBefore);
        
        try {
            const response = await fetch('/validate_move', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ fen: fenBefore, move: move })
            });

            if (!response.ok) {
                throw new Error(`Server Error: ${response.status}`);
            }

            const result = await response.json();
            console.log('Validation result:', result);

            if (result.valid) {
                console.log('Move is valid! Executing...');
                
                // Make the move on the board
                makeMove(move);
                
                // Log the new state
                const fenAfter = boardToFen();
                console.log('FEN after move:', fenAfter);
                
                // Deselect piece
                fromSquare = null;
                
                // Update UI immediately
                updateUi();
                
                // Get AI response after a short delay
                setTimeout(() => {
                    console.log('Requesting AI move...');
                    getAiMove();
                }, 300);
            } else {
                console.log('Move is invalid!');
                statusElement.textContent = 'Invalid Move! Try again.';
                isAwaitingAi = false;
                updateUi();
            }
            
        } catch (error) {
            console.error('Error validating move:', error);
            statusElement.textContent = 'Error validating move: ' + error.message;
            isAwaitingAi = false;
            fromSquare = null;
            updateUi();
        }
    }

    async function getAiMove() {
        const fen = boardToFen();
        console.log('Requesting AI move for FEN:', fen);
        
        try {
            const response = await fetch('/get_move', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ fen: fen })
            });

            if (!response.ok) {
                throw new Error(`Server Error: ${response.statusText}`);
            }

            const aiMove = await response.json();
            console.log('AI move received:', aiMove);
            
            if (aiMove && aiMove.FromRow !== undefined) {
                makeMove(aiMove);
                const fenAfter = boardToFen();
                console.log('FEN after AI move:', fenAfter);
            } else {
                console.log('No AI move - game over?');
                statusElement.textContent = 'Game Over!';
            }

        } catch (error) {
            console.error('Error getting AI move:', error);
            statusElement.textContent = 'Error communicating with engine: ' + error.message;
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

                // Highlight selected square
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
        if (isAwaitingAi) {
            statusElement.textContent = 'Black is thinking...';
        } else {
            statusElement.textContent = 'White to move';
        }
    }

    function makeMove(move) {
        console.log('makeMove called with:', move);
        console.log('Board before:', JSON.stringify(boardState));
        
        const piece = boardState[move.FromRow][move.FromCol];
        console.log('Moving piece:', piece, 'from', move.FromRow, move.FromCol, 'to', move.ToRow, move.ToCol);
        
        boardState[move.ToRow][move.ToCol] = piece;
        boardState[move.FromRow][move.FromCol] = ' ';
        
        console.log('Board after:', JSON.stringify(boardState));
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
        fromSquare = null;
        isAwaitingAi = false;
        statusElement.textContent = 'White to move';
        updateUi();
        console.log('Game initialized');
    }
    
    restartButton.addEventListener('click', initGame);
    initGame();
});