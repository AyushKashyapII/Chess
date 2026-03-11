//go:build js && wasm

package main

import (
	"chess-engine/handlers"
	//"encoding/json"
	"fmt"
	//"log"
	//"net/http"
	//"os"
	"strings"
	"syscall/js"
)

var currentBoard [8][8]rune

type MoveRequest struct {
	Fen string `json:"fen"`
}

type Move struct {
	FromRow int `json:"FromRow"`
	FromCol int `json:"FromCol"`
	ToRow   int `json:"ToRow"`
	ToCol   int `json:"ToCol"`
}

type ValidateRequest struct {
	//Fen  string `json:"fen"`
	Move Move   `json:"move"`
}

type ValidateResponse struct {
	GameStatus bool `json:"gamestatus"`
	Valid  bool   `json:"valid"`
	NewFen string `json:"newFen,omitempty"`
}

func main() {
	handlers.InitZobrist()

	// Register JS functions
	js.Global().Set("init_board_wasm", js.FuncOf(init_board_wasm))
	js.Global().Set("validate_move_wasm", js.FuncOf(validate_move_wasm))
	js.Global().Set("get_ai_move_wasm", js.FuncOf(get_ai_move_wasm))

	// Keep the Go program running
	c := make(chan struct{}, 0)
	<-c
}

func init_board_wasm(this js.Value, args []js.Value) interface{} {
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"
	if len(args) > 0 {
		fen = args[0].String()
	}
	currentBoard = parseFEN(fen)
	return nil
}

func get_ai_move_wasm(this js.Value, args []js.Value) interface{} {
	bestMove := handlers.FindBestMove(currentBoard, false)

	if bestMove.FromRow == 0 && bestMove.FromCol == 0 &&
		bestMove.ToRow == 0 && bestMove.ToCol == 0 {
		return js.ValueOf(map[string]interface{}{
			"valid":  false,
			"newFen": boardToFEN(currentBoard),
		})
	}

	move := Move{
		FromRow: bestMove.FromRow,
		FromCol: bestMove.FromCol,
		ToRow:   bestMove.ToRow,
		ToCol:   bestMove.ToCol,
	}

	currentBoard = applyMove(currentBoard, move)

	isPossibleMove := handlers.FindBestMove(currentBoard, true)
	isPossible := true
	if isPossibleMove.FromRow == 0 && isPossibleMove.FromCol == 0 &&
		isPossibleMove.ToRow == 0 && isPossibleMove.ToCol == 0 {
		isPossible = false
	}

	return js.ValueOf(map[string]interface{}{
		"gamestatus": isPossible,
		"valid":      true,
		"fromR":      move.FromRow,
		"fromC":      move.FromCol,
		"toR":        move.ToRow,
		"toC":        move.ToCol,
		"newFen":     boardToFEN(currentBoard),
	})
}

func validate_move_wasm(this js.Value, args []js.Value) interface{} {
	if len(args) < 4 {
		return js.ValueOf(map[string]interface{}{"error": "missing arguments"})
	}

	fromRow := args[0].Int()
	fromCol := args[1].Int()
	toRow := args[2].Int()
	toCol := args[3].Int()

	valid := handlers.IsValidMove(
		currentBoard,
		currentBoard[fromRow][fromCol],
		fromRow,
		fromCol,
		toRow,
		toCol,
		nil,
	)

	if valid {
		move := Move{FromRow: fromRow, FromCol: fromCol, ToRow: toRow, ToCol: toCol}
		currentBoard = applyMove(currentBoard, move)
		return js.ValueOf(map[string]interface{}{
			"valid":      true,
			"newFen":     boardToFEN(currentBoard),
			"gamestatus": true,
		})
	}

	return js.ValueOf(map[string]interface{}{
		"valid": false,
	})
}

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

func applyMove(board [8][8]rune, move Move) [8][8]rune {
	var newBoard [8][8]rune
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			newBoard[i][j] = board[i][j]
		}
	}

	piece := board[move.FromRow][move.FromCol]
	newBoard[move.FromRow][move.FromCol] = 0

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
	if (piece == 'P' || piece == 'p') && move.ToCol != move.FromCol {
		if board[move.ToRow][move.ToCol] == 0 {
			newBoard[move.FromRow][move.ToCol] = 0
		}
	}
	newBoard[move.ToRow][move.ToCol] = piece
	if piece == 'P' && move.ToRow == 0 {
		newBoard[move.ToRow][move.ToCol] = 'Q'
	} else if piece == 'p' && move.ToRow == 7 {
		newBoard[move.ToRow][move.ToCol] = 'q'
	}

	return newBoard
}

func boardToFEN(board [8][8]rune) string {
	var fen strings.Builder

	for row := 0; row < 8; row++ {
		emptyCount := 0

		for col := 0; col < 8; col++ {
			piece := board[row][col]

			if piece == 0 {
				emptyCount++
			} else {
				if emptyCount > 0 {
					fen.WriteString(fmt.Sprintf("%d", emptyCount))
					emptyCount = 0
				}
				fen.WriteRune(piece)
			}
		}

		if emptyCount > 0 {
			fen.WriteString(fmt.Sprintf("%d", emptyCount))
		}

		if row < 7 {
			fen.WriteString("/")
		}
	}

	return fen.String()
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
