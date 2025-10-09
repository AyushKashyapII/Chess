package handlers

import "fmt"

var PieceValues = map[rune]int{
	'p': 10,
	'P': 10,
	'n': 30,
	'N': 30,
	'b': 30,
	'B': 30,
	'r': 50,
	'R': 50,
	'q': 90,
	'Q': 90,
	'k': 900,
	'K': 900,
}

func bestMove(board [8][8]rune) {
	fmt.Println("best possible move")
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if board[i][j] != 0 {
				fmt.Println("Piece found")
				if isWhite(board[i][j]) {
					fmt.Println("White piece found")
					fmt.Println(GetValue(board[i][j]))
				}
			}
		}
	}

}

// GetValue returns the value of a given piece
func GetValue(piece rune) int {
	fmt.Println("GetValue")
	return PieceValues[piece]
}
