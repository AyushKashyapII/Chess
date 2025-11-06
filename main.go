package main

import (
	"chess-engine/handlers"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

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
	Fen  string `json:"fen"`
	Move Move   `json:"move"`
}

type ValidateResponse struct {
	Valid  bool   `json:"valid"`
	NewFen string `json:"newFen,omitempty"`
}

func main() {
	fmt.Println("CHESS ENGINE!!!")
	handlers.InitZobrist()
	http.HandleFunc("/get_move", handleGetMove)
	http.HandleFunc("/validate_move", validate_move)
	http.Handle("/", http.FileServer(http.Dir("./frontend")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting the chess engine at port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleGetMove(w http.ResponseWriter, r *http.Request) {
	var request MoveRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error: Invalid request format", http.StatusBadRequest)
		return
	}

	fmt.Println("Received FEN from client:", request.Fen)
	board := parseFEN(request.Fen)
	// for i := 0; i < 8; i++ {
	// 	for j := 0; j < 8; j++ {
	// 		if board[i][j] == 0 {
	// 			fmt.Print(".")
	// 		} else {
	// 			fmt.Print(string(board[i][j]))
	// 		}
	// 	}
	// 	fmt.Println()
	// }

	bestMove := handlers.FindBestMove(board, false)
	fmt.Printf("Engine chose to move from (%d,%d) to (%d,%d)\n",
		bestMove.FromRow, bestMove.FromCol, bestMove.ToRow, bestMove.ToCol)
	if bestMove.FromRow == 0 && bestMove.FromCol == 0 &&
		bestMove.ToRow == 0 && bestMove.ToCol == 0 {
		fmt.Println("WARNING: AI returned no move (all zeros)")
		response := ValidateResponse{
			Valid:  false,
			NewFen: request.Fen,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
	aiMove := Move{
		FromRow: bestMove.FromRow,
		FromCol: bestMove.FromCol,
		ToRow:   bestMove.ToRow,
		ToCol:   bestMove.ToCol,
	}
	fmt.Println("ai move response", aiMove)
	newBoard := applyMove(board, aiMove)

	
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if newBoard[i][j] == 0 || newBoard[i][j] == ' ' {
				fmt.Print(". ")
			} else {
				fmt.Printf("%c ", newBoard[i][j])
			}
		}
		fmt.Println("")
	}


	response := ValidateResponse{
		Valid:  true,
		NewFen: boardToFEN(newBoard),
	}

	fmt.Println("AI move new FEN:", response.NewFen)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func validate_move(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var data ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	fmt.Println("request validate move:", data.Move)
	fmt.Println("request fen:", data.Fen)

	parsedBoard := parseFEN(data.Fen)
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			fmt.Print(string(parsedBoard[i][j]))
		}
		fmt.Println()
	}
	valid := handlers.IsValidMove(
		parsedBoard,
		parsedBoard[data.Move.FromRow][data.Move.FromCol],
		data.Move.FromRow,
		data.Move.FromCol,
		data.Move.ToRow,
		data.Move.ToCol,
		nil,
	)

	fmt.Println("valid ", valid)
	//fmt.Println("fen rece",data.Fen)

	response := ValidateResponse{
		Valid:  valid,
		NewFen: data.Fen,
	}

	if valid {
		newBoard := applyMove(parsedBoard, data.Move)
		response.NewFen = boardToFEN(newBoard)
		fmt.Println("Move is valid!")
		fmt.Println("new FEN:", response.NewFen)
	} else {
		fmt.Println("Move is INVALID")
	}

	fmt.Println("Sending response:", response)
	json.NewEncoder(w).Encode(response)
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
	newBoard[move.FromRow][move.FromCol] = ' '
	
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
		if board[move.ToRow][move.ToCol] == ' ' || board[move.ToRow][move.ToCol] == 0 {
			newBoard[move.FromRow][move.ToCol] = ' '
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

			if piece == ' ' || piece == 0 {
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
