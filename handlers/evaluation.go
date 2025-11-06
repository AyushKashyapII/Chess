package handlers

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

// type Move struct {
// 	FromRow,FromCol int
// 	ToRow,ToCol int
// }

type HashMap struct {
	HashKey  uint64
	Score    int
	Depth    int
	BestMove Move
}

const ttSize = 512

var transpositionTable [ttSize]HashMap

var WhitePawnPST = [8][8]int{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{50, 50, 50, 50, 50, 50, 50, 50},
	{10, 10, 20, 30, 30, 20, 10, 10},
	{5, 5, 10, 25, 25, 10, 5, 5},
	{0, 0, 0, 20, 20, 0, 0, 0},
	{5, -5, -10, 0, 0, -10, -5, 5},
	{5, 10, 10, -20, -20, 10, 10, 5},
	{-30, -30, -30, -30, -30, -30, -30, -30},
}

var WhiteKnightPST = [8][8]int{
	{-50, -40, -30, -30, -30, -30, -40, -50},
	{-40, -20, 0, 0, 0, 0, -20, -40},
	{-30, 0, 10, 15, 15, 10, 0, -30},
	{-30, 5, 15, 20, 20, 15, 5, -30},
	{-30, 0, 15, 20, 20, 15, 0, -30},
	{-30, 5, 10, 15, 15, 10, 5, -30},
	{-40, -20, 0, 5, 5, 0, -20, -40},
	{-50, -40, -30, -30, -30, -30, -40, -50},
}

var WhiteBishopPST = [8][8]int{
	{-20, -10, -10, -10, -10, -10, -10, -20},
	{-10, 0, 0, 0, 0, 0, 0, -10},
	{-10, 0, 5, 10, 10, 5, 0, -10},
	{-10, 5, 5, 10, 10, 5, 5, -10},
	{-10, 0, 10, 10, 10, 10, 0, -10},
	{-10, 10, 10, 10, 10, 10, 10, -10},
	{-10, 5, 0, 0, 0, 0, 5, -10},
	{-20, -10, -10, -10, -10, -10, -10, -20},
}

var WhiteRookPST = [8][8]int{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{5, 10, 10, 10, 10, 10, 10, 5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{0, 0, 0, 5, 5, 0, 0, 0},
}

var WhiteQueenPST = [8][8]int{
	{-20, -10, -10, -5, -5, -10, -10, -20},
	{-10, 0, 0, 0, 0, 0, 0, -10},
	{-10, 0, 5, 5, 5, 5, 0, -10},
	{-5, 0, 5, 5, 5, 5, 0, -5},
	{0, 0, 5, 5, 5, 5, 0, -5},
	{-10, 5, 5, 5, 5, 5, 0, -10},
	{-10, 0, 5, 0, 0, 0, 0, -10},
	{-20, -10, -10, -5, -5, -10, -10, -20},
}
var PieceValues = map[rune]int{
	'p': -10,
	'P': 10,
	'n': -30,
	'N': 30,
	'b': -30,
	'B': 30,
	'r': -50,
	'R': 50,
	'q': -90,
	'Q': 90,
	'k': -900,
	'K': 900,
}
var WhiteKingMiddlegamePST = [8][8]int{
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-20, -30, -30, -40, -40, -30, -30, -20},
	{-10, -20, -20, -20, -20, -20, -20, -10},
	{20, 20, 0, 0, 0, 0, 20, 20},
	{20, 30, 10, 0, 0, 10, 30, 20},
}
var WhiteKingEndgamePST = [8][8]int{
	{-50, -40, -30, -20, -20, -30, -40, -50},
	{-30, -20, -10, 0, 0, -10, -20, -30},
	{-30, -10, 20, 30, 30, 20, -10, -30},
	{-30, -10, 30, 40, 40, 30, -10, -30},
	{-30, -10, 30, 40, 40, 30, -10, -30},
	{-30, -10, 20, 30, 30, 20, -10, -30},
	{-30, -30, 0, 0, 0, 0, -30, -30},
	{-50, -30, -30, -30, -30, -30, -30, -50},
}
var zobristTable [12][64]uint64
var current_hash uint64

func randomUnit64() uint64 {
	var buf [8]byte
	_, _ = rand.Read(buf[:])
	return binary.LittleEndian.Uint64(buf[:])
}

func InitZobrist() {
	for p := 0; p < 12; p++ {
		for sq := 0; sq < 64; sq++ {
			zobristTable[p][sq] = randomUnit64()
		}
	}
	fmt.Println("Zobrist Table Initialised!!!")
}

var pieceToIndex = map[rune]int{
	'P': 0, 'N': 1, 'B': 2, 'R': 3, 'Q': 4, 'K': 5,
	'p': 6, 'n': 7, 'b': 8, 'r': 9, 'q': 10, 'k': 11,
}

func GetZobristValue(board [8][8]rune) uint64 {
	var hash uint64 = 0
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			piece := board[row][col]
			if pieceToIndex[piece] >= 0 {
				pieceIndex := pieceToIndex[piece]
				squareIndex := row*8 + col
				hash ^= zobristTable[pieceIndex][squareIndex]
				//continue
			}

		}
	}
	current_hash = hash
	return hash
}

func GetCurrentHash() uint64 {
	return current_hash
}

func UpdateHashForMove(currentHash uint64, move Move, board [8][8]rune) uint64 {
	newHash := currentHash

	fromPiece := board[move.FromRow][move.FromCol]
	if fromPiece != '.' && fromPiece != 0 {
		pieceIdx := pieceToIndex[fromPiece]
		fromSquare := move.FromRow*8 + move.FromCol
		newHash ^= zobristTable[pieceIdx][fromSquare]
	}

	toPiece := board[move.ToRow][move.ToCol]
	if toPiece != '.' && toPiece != 0 {
		pieceIdx := pieceToIndex[toPiece]
		toSquare := move.ToRow*8 + move.ToCol
		newHash ^= zobristTable[pieceIdx][toSquare]
	}

	if fromPiece != '.' && fromPiece != 0 {
		pieceIdx := pieceToIndex[fromPiece]
		toSquare := move.ToRow*8 + move.ToCol
		newHash ^= zobristTable[pieceIdx][toSquare]
	}

	return newHash
}

func GetValue(piece rune) int {
	return PieceValues[piece]
}

func Evaluate_board(board [8][8]rune) int {
	board_state := 0
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece := board[i][j]
			if piece != 0 {
				board_state += GetValue(board[i][j])
				switch piece {
				case 'P':
					board_state += WhitePawnPST[i][j]
				case 'p':
					board_state -= WhitePawnPST[7-i][j]
				case 'Q':
					board_state += WhiteQueenPST[i][j]
				case 'q':
					board_state -= WhiteQueenPST[7-i][j]
				case 'N':
					board_state += WhiteKnightPST[i][j]
				case 'n':
					board_state -= WhiteKnightPST[7-i][j]
				case 'B':
					board_state += WhiteBishopPST[i][j]
				case 'b':
					board_state -= WhiteBishopPST[7-i][j]
				case 'R':
					board_state += WhiteRookPST[i][j]
				case 'r':
					board_state -= WhiteRookPST[7-i][j]
				case 'K':
					board_state += WhiteKingMiddlegamePST[i][j]
				case 'k':
					board_state -= WhiteKingMiddlegamePST[7-i][j]
				}
			}
		}
	}
	return board_state
}
