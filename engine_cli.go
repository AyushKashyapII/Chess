//go:build !js

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"chess-engine/handlers"
)

// parseFEN converts a simple piece-placement FEN (first field) into a board.
func parseFEN(fen string) [8][8]rune {
	var board [8][8]rune
	rows := strings.Split(fen, "/")
	for rowIdx, row := range rows {
		colIdx := 0
		for _, char := range row {
			if char >= '1' && char <= '8' {
				colIdx += int(char - '0')
			} else {
				board[rowIdx][colIdx] = char
				colIdx++
			}
		}
	}
	return board
}

// boardToFEN converts a board back to the first FEN field (piece placement only).
func boardToFEN(board [8][8]rune) string {
	var sb strings.Builder
	for row := 0; row < 8; row++ {
		empty := 0
		for col := 0; col < 8; col++ {
			p := board[row][col]
			if p == 0 {
				empty++
			} else {
				if empty > 0 {
					sb.WriteString(fmt.Sprintf("%d", empty))
					empty = 0
				}
				sb.WriteRune(p)
			}
		}
		if empty > 0 {
			sb.WriteString(fmt.Sprintf("%d", empty))
		}
		if row < 7 {
			sb.WriteString("/")
		}
	}
	return sb.String()
}

// printBoard draws the board in the terminal.
func printBoard(board [8][8]rune) {
	fmt.Println()
	for row := 0; row < 8; row++ {
		fmt.Printf("%d ", 8-row)
		for col := 0; col < 8; col++ {
			p := board[row][col]
			if p == 0 {
				fmt.Print(". ")
			} else {
				fmt.Printf("%c ", p)
			}
		}
		fmt.Println()
	}
	fmt.Println("  a b c d e f g h")
	fmt.Println()
}

// algebraic square like "e2" -> board coordinates.
func squareToCoords(s string) (row, col int, ok bool) {
	if len(s) != 2 {
		return 0, 0, false
	}
	file := s[0]
	rank := s[1]
	if file < 'a' || file > 'h' || rank < '1' || rank > '8' {
		return 0, 0, false
	}
	col = int(file - 'a')
	// internal row 0 is rank 8, row 7 is rank 1
	row = 8 - int(rank-'0')
	return row, col, true
}

// coords -> "e2"
func coordsToSquare(row, col int) string {
	file := 'a' + rune(col)
	rank := '8' - rune(row)
	return fmt.Sprintf("%c%c", file, rank)
}

// simple abs helper.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// applyMove copies the board and applies a move (with basic castling/en-passant/promotion).
func applyMove(board [8][8]rune, move handlers.Move) [8][8]rune {
	var newBoard [8][8]rune
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			newBoard[i][j] = board[i][j]
		}
	}

	piece := board[move.FromRow][move.FromCol]
	newBoard[move.FromRow][move.FromCol] = 0

	// Castling rook move
	if piece == 'K' || piece == 'k' {
		if abs(move.ToCol-move.FromCol) == 2 {
			if move.ToCol > move.FromCol {
				newBoard[move.FromRow][5] = newBoard[move.FromRow][7]
				newBoard[move.FromRow][7] = ' '
			} else {
				newBoard[move.FromRow][3] = newBoard[move.FromRow][0]
				newBoard[move.FromRow][0] = ' '
			}
		}
	}
	// En-passant capture
	if (piece == 'P' || piece == 'p') && move.ToCol != move.FromCol {
		if board[move.ToRow][move.ToCol] == 0 {
			newBoard[move.FromRow][move.ToCol] = 0
		}
	}

	newBoard[move.ToRow][move.ToCol] = piece
	// Promotion
	if piece == 'P' && move.ToRow == 0 {
		newBoard[move.ToRow][move.ToCol] = 'Q'
	} else if piece == 'p' && move.ToRow == 7 {
		newBoard[move.ToRow][move.ToCol] = 'q'
	}

	return newBoard
}

func isWhite(piece rune) bool {
	return piece == 'P' || piece == 'N' || piece == 'B' || piece == 'R' || piece == 'Q' || piece == 'K'
}

func main() {
	handlers.InitZobrist()
	reader := bufio.NewReader(os.Stdin)

	startFen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"

	fmt.Println("=== Terminal Chess Engine (you are White, engine is Black) ===")
	fmt.Println("Enter FEN (piece placement only) or press Enter for the normal start position:")
	fmt.Printf("FEN [%s]: ", startFen)

	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		line = startFen
	}

	board := parseFEN(line)
	whiteToMove := true // you start as White

	for {
		printBoard(board)

		if whiteToMove {
			fmt.Println("Your move (format: e2e4, or 'q' to quit):")
			fmt.Print("> ")
			moveStr, _ := reader.ReadString('\n')
			moveStr = strings.TrimSpace(strings.ToLower(moveStr))

			if moveStr == "q" || moveStr == "quit" || moveStr == "exit" {
				fmt.Println("Exiting game.")
				return
			}

			if len(moveStr) != 4 {
				fmt.Println("Invalid input. Use format like e2e4.")
				continue
			}

			fromSq := moveStr[0:2]
			toSq := moveStr[2:4]

			fromRow, fromCol, ok1 := squareToCoords(fromSq)
			toRow, toCol, ok2 := squareToCoords(toSq)
			if !ok1 || !ok2 {
				fmt.Println("Invalid squares. Use a1..h8.")
				continue
			}

			piece := board[fromRow][fromCol]
			if piece == 0 {
				fmt.Println("No piece on that square.")
				continue
			}
			if !isWhite(piece) {
				fmt.Println("That's not your (White) piece.")
				continue
			}

			var promotionPiece *rune
			if piece == 'P' && toRow == 0 {
				q := 'Q'
				promotionPiece = &q
			}

			if !handlers.IsValidMove(board, piece, fromRow, fromCol, toRow, toCol, promotionPiece) {
				fmt.Println("Illegal move according to engine rules.")
				continue
			}

			mv := handlers.Move{FromRow: fromRow, FromCol: fromCol, ToRow: toRow, ToCol: toCol}
			board = applyMove(board, mv)
			whiteToMove = false

		} else {
			fmt.Println("Engine thinking...")
			start := time.Now()
			bestMove := handlers.FindBestMove(board, whiteToMove)
			elapsed := time.Since(start)

			if bestMove.FromRow == 0 && bestMove.FromCol == 0 &&
				bestMove.ToRow == 0 && bestMove.ToCol == 0 {
				fmt.Println("Engine has no legal moves. Game over.")
				return
			}

			board = applyMove(board, bestMove)
			fmt.Printf("Engine plays: %s%s (took %v)\n",
				coordsToSquare(bestMove.FromRow, bestMove.FromCol),
				coordsToSquare(bestMove.ToRow, bestMove.ToCol),
				elapsed)

			fmt.Println("New FEN (piece placement):", boardToFEN(board))
			whiteToMove = true
		}
	}
}
