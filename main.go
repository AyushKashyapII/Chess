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

func main() {
	fmt.Println(("CHESS ENGINE!!!"))
	handlers.InitZobrist()
	http.HandleFunc("/get_move", handleGetMove)
	http.HandleFunc("/validate_move", validate_move)
	http.Handle("/", http.FileServer(http.Dir("./frontend")))

	port := os.Getenv(("PORT"))
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting the chess engine at port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
	fmt.Println("getting here ")

}

func handleGetMove(w http.ResponseWriter, r *http.Request) {
	var request MoveRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error: Invalid request formayt ", http.StatusBadRequest)
		return
	}

	fmt.Println("Recievdd FEN from client:", request.Fen)
	board := parseFEN(request.Fen)
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			fmt.Print(string(board[i][j]))
		}
		fmt.Println()
	}
	//generatetry(board,false)

	bestMove := handlers.FindBestMove(board, false)
	fmt.Printf("Engine chose to move from (%d,%d) to (%d,%d)", bestMove.FromRow, bestMove.FromCol, bestMove.ToRow, bestMove.ToCol)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bestMove); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

}

func validate_move(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var data ValidateRequest
	json.NewDecoder(r.Body).Decode(&data)
	fmt.Println("request validate move:", data.Move)
	fmt.Println("request fen ", data.Fen)
	parsedBoard := parseFEN(data.Fen)
	valid := handlers.IsValidMove(parsedBoard, parsedBoard[data.Move.FromRow][data.Move.FromCol], data.Move.FromRow, data.Move.FromCol, data.Move.ToRow, data.Move.ToCol, nil)

	json.NewEncoder(w).Encode(map[string]bool{"valid": valid})
}

// func generatetry(board [8][8] rune,isWhiteTurn bool) {
// 	if isWhiteTurn {
// 		fmt.Print("WHite tuen ")
// 	}else{
// 		fmt.Print(("Black turn "))
// 	}
// 	for i:=0;i<8;i++{
// 		for j:=0;j<8;j++{
// 			var piec=board[i][j]
// 			if piec=="."{
// 				fmt.Print(" ")
// 			}else{
// 				fmt.Print(string(piec))
// 			}
// 		}
// 	}
// }

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
