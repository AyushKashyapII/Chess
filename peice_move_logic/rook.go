package peice_move_logic


func GetRookMoves(board [8][8]rune, x, y int) []Move {
	var moves []Move

	// Rook can move in 4 directions
	// 1. Up
	// 2. Down
	// 3. Left
	// 4. Right
	// Rook can move any number of squares in any direction
	// Rook can't jump over other pieces

	directions := []struct{ dx, dy int }{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

	for _, d := range directions {
		for i := 1; i < 8; i++ {
			newX, newY := x+d.dx*i, y+d.dy*i

			if newX < 0 || newX >= 8 || newY < 0 || newY >= 8 || board[newX][newY] == ' ' {
				break
			}

			moves = append(moves, Move{X: newX, Y: newY})

			if board[newX][newY] != ' ' {
				break
			}
		}
	}

	return moves
}
