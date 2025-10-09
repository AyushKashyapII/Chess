package peice_move_logic

type Piece interface {
	GetValidMoves(board [8][8]rune, x, y int) []Move
}

type Move struct {
	X, Y int
}
