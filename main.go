//go:build js && wasm

package main

import (
	"chess-engine/handlers"
	"encoding/json"
	"fmt"
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

	js.Global().Set("init_board_wasm", js.FuncOf(init_board_wasm))
	js.Global().Set("validate_move_string_wasm", js.FuncOf(validate_move_string_wasm))
	js.Global().Set("get_ai_move_string_wasm", js.FuncOf(get_ai_move_string_wasm))
	
	// Root splitting functions for parallel search
	js.Global().Set("get_all_legal_moves_wasm", js.FuncOf(get_all_legal_moves_wasm))
	js.Global().Set("search_subset_wasm", js.FuncOf(search_subset_wasm))
	js.Global().Set("apply_move_wasm", js.FuncOf(apply_move_wasm))
	
	// Keep old functions for backward compatibility
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
				newBoard[move.FromRow][7] = 0
			} else {
				newBoard[move.FromRow][3] = newBoard[move.FromRow][0]
				newBoard[move.FromRow][0] = 0
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

// squareToCoords converts algebraic square like "e2" -> board coordinates (like engine_cli.go)
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

// coordsToSquare converts board coordinates -> "e2" (like engine_cli.go)
func coordsToSquare(row, col int) string {
	file := 'a' + rune(col)
	rank := '8' - rune(row)
	return fmt.Sprintf("%c%c", file, rank)
}

// isWhite checks if a piece is white (like engine_cli.go)
func isWhitePiece(piece rune) bool {
	return piece == 'P' || piece == 'N' || piece == 'B' || piece == 'R' || piece == 'Q' || piece == 'K'
}

// validate_move_string_wasm accepts move in format "e2e4" (like engine_cli.go)
func validate_move_string_wasm(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{"valid": false, "error": "missing move string"})
	}

	moveStr := strings.TrimSpace(strings.ToLower(args[0].String()))
	
	if len(moveStr) != 4 {
		return js.ValueOf(map[string]interface{}{"valid": false, "error": "invalid move format, use e2e4"})
	}

	fromSq := moveStr[0:2]
	toSq := moveStr[2:4]

	fromRow, fromCol, ok1 := squareToCoords(fromSq)
	toRow, toCol, ok2 := squareToCoords(toSq)
	
	if !ok1 || !ok2 {
		return js.ValueOf(map[string]interface{}{"valid": false, "error": "invalid squares"})
	}

	piece := currentBoard[fromRow][fromCol]
	if piece == 0 {
		return js.ValueOf(map[string]interface{}{"valid": false, "error": "no piece on that square"})
	}
	
	if !isWhitePiece(piece) {
		return js.ValueOf(map[string]interface{}{"valid": false, "error": "not a white piece"})
	}

	var promotionPiece *rune
	if piece == 'P' && toRow == 0 {
		q := 'Q'
		promotionPiece = &q
	}

	valid := handlers.IsValidMove(
		currentBoard,
		piece,
		fromRow,
		fromCol,
		toRow,
		toCol,
		promotionPiece,
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
		"error": "illegal move",
	})
}

// get_ai_move_string_wasm returns AI move in format "e2e4" (like engine_cli.go)
func get_ai_move_string_wasm(this js.Value, args []js.Value) interface{} {
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

	// Check if white has any moves left
	isPossibleMove := handlers.FindBestMove(currentBoard, true)
	isPossible := true
	if isPossibleMove.FromRow == 0 && isPossibleMove.FromCol == 0 &&
		isPossibleMove.ToRow == 0 && isPossibleMove.ToCol == 0 {
		isPossible = false
	}

	// Convert to string format like "e2e4"
	moveString := coordsToSquare(move.FromRow, move.FromCol) + coordsToSquare(move.ToRow, move.ToCol)

	return js.ValueOf(map[string]interface{}{
		"gamestatus": isPossible,
		"valid":      true,
		"move":       moveString,
		"newFen":     boardToFEN(currentBoard),
	})
}

// get_all_legal_moves_wasm returns all legal moves as JSON string for the given FEN
func get_all_legal_moves_wasm(this js.Value, args []js.Value) interface{} {
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"
	if len(args) > 0 {
		fen = args[0].String()
	}
	
	board := parseFEN(fen)
	isWhiteTurn := false // For AI move, it's black's turn
	if len(args) > 1 {
		isWhiteTurn = args[1].Bool()
	}
	
	allMoves := handlers.GenereateAllMoves(board, isWhiteTurn)
	
	// Convert to JSON-serializable format
	type MoveJSON struct {
		FromRow int `json:"fromRow"`
		FromCol int `json:"fromCol"`
		ToRow   int `json:"toRow"`
		ToCol   int `json:"toCol"`
	}
	
	movesJSON := make([]MoveJSON, len(allMoves))
	for i, move := range allMoves {
		movesJSON[i] = MoveJSON{
			FromRow: move.FromRow,
			FromCol: move.FromCol,
			ToRow:   move.ToRow,
			ToCol:   move.ToCol,
		}
	}
	
	jsonBytes, err := json.Marshal(movesJSON)
	if err != nil {
		return js.ValueOf(map[string]interface{}{"error": err.Error()})
	}
	
	return js.ValueOf(string(jsonBytes))
}

// search_subset_wasm searches only the provided moves and returns best move and score
func search_subset_wasm(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf(map[string]interface{}{"error": "missing arguments"})
	}
	
	fen := args[0].String()
	movesJson := args[1].String()
	
	board := parseFEN(fen)
	isWhiteTurn := false // For AI move, it's black's turn
	
	// Parse moves from JSON
	type MoveJSON struct {
		FromRow int `json:"fromRow"`
		FromCol int `json:"fromCol"`
		ToRow   int `json:"toRow"`
		ToCol   int `json:"toCol"`
	}
	
	var movesJSON []MoveJSON
	if err := json.Unmarshal([]byte(movesJson), &movesJSON); err != nil {
		return js.ValueOf(map[string]interface{}{"error": "invalid moves JSON: " + err.Error()})
	}
	
	// Convert to handlers.Move
	movesToSearch := make([]handlers.Move, len(movesJSON))
	for i, m := range movesJSON {
		movesToSearch[i] = handlers.Move{
			FromRow: m.FromRow,
			FromCol: m.FromCol,
			ToRow:   m.ToRow,
			ToCol:   m.ToCol,
		}
	}
	
	// Search the subset
	bestMove, bestScore := handlers.SearchSpecificMoves(board, isWhiteTurn, movesToSearch)
	
	// Convert move to string format
	moveString := coordsToSquare(bestMove.FromRow, bestMove.FromCol) + coordsToSquare(bestMove.ToRow, bestMove.ToCol)
	
	return js.ValueOf(map[string]interface{}{
		"move":  moveString,
		"score": bestScore,
		"fromRow": bestMove.FromRow,
		"fromCol": bestMove.FromCol,
		"toRow":   bestMove.ToRow,
		"toCol":   bestMove.ToCol,
	})
}

// apply_move_wasm applies a move to the board and returns the new FEN
func apply_move_wasm(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf(map[string]interface{}{"error": "missing arguments"})
	}
	
	fen := args[0].String()
	moveJson := args[1].String()
	
	board := parseFEN(fen)
	
	// Parse move from JSON
	type MoveJSON struct {
		FromRow int `json:"fromRow"`
		FromCol int `json:"fromCol"`
		ToRow   int `json:"toRow"`
		ToCol   int `json:"toCol"`
	}
	
	var moveJSON MoveJSON
	if err := json.Unmarshal([]byte(moveJson), &moveJSON); err != nil {
		return js.ValueOf(map[string]interface{}{"error": "invalid move JSON: " + err.Error()})
	}
	
	move := Move{
		FromRow: moveJSON.FromRow,
		FromCol: moveJSON.FromCol,
		ToRow:   moveJSON.ToRow,
		ToCol:   moveJSON.ToCol,
	}
	
	// Apply move using the proper Go function
	newBoard := applyMove(board, move)
	newFen := boardToFEN(newBoard)
	
	return js.ValueOf(map[string]interface{}{
		"newFen": newFen,
	})
}
