package handlers

import "fmt"

type CastlingRights struct {
	WhiteKingSide  bool
	WhiteQueenSide bool
	BlackKingSide  bool
	BlackQueenSide bool
}

type Move struct {
	FromRow,FromCol int
	ToRow,ToCol int
}

var initialPositions = map[string]bool{
	"e1": true,
	"e8": true, 
	"a1": true,
	"h1": true, 
	"a8": true, 
	"h8": true, 
}

func UpdateCastlingRights(board [8][8]rune, fromRow, fromCol int, castlingRights *CastlingRights) {
	piece := board[fromRow][fromCol]

	switch piece {
	case 'K':
		castlingRights.WhiteKingSide = false
		castlingRights.WhiteQueenSide = false
	case 'k':
		castlingRights.BlackKingSide = false
		castlingRights.BlackQueenSide = false
	case 'R':
		if fromRow == 7 && fromCol == 0 {
			castlingRights.WhiteQueenSide = false
		} else if fromRow == 7 && fromCol == 7 {
			castlingRights.WhiteKingSide = false
		}
	case 'r':
		if fromRow == 0 && fromCol == 0 {
			castlingRights.BlackQueenSide = false
		} else if fromRow == 0 && fromCol == 7 {
			castlingRights.BlackKingSide = false
		}
	}
}

func IsSquareUnderAttack(board [8][8]rune, row, col int, isWhitePiece bool) bool {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece := board[i][j]
			if piece == 0 || isWhite(piece) == isWhitePiece {
				continue
			}
			if IsValidMove(board, piece, i, j, row, col, nil) {
				return true
			}
		}
	}
	return false
}

func IsInCheck(board [8][8]rune, isWhiteKing bool, kingRow, kingCol int) bool {
	return IsSquareUnderAttack(board, kingRow, kingCol, isWhiteKing)
}

func IsCastleable(board [8][8]rune, fromRow, fromCol, toRow, toCol int) bool {
	piece := board[fromRow][fromCol]

	if (piece != 'K' && piece != 'k') || abs(fromCol-toCol) != 2 || fromRow != toRow {
		return false
	}

	isKingSide := toCol > fromCol
	row := fromRow
	isWhiteKing := piece == 'K'

	if IsInCheck(board, isWhiteKing, row, fromCol) {
		return false
	}

	if isKingSide {
		rookCol := 7
		if (isWhiteKing && board[7][rookCol] != 'R') || (!isWhiteKing && board[0][rookCol] != 'r') {
			return false
		}
		for col := fromCol + 1; col < rookCol; col++ {
			if board[row][col] != 0 || IsSquareUnderAttack(board, row, col, isWhiteKing) {
				return false
			}
		}
	} else {
		rookCol := 0
		if (isWhiteKing && board[7][rookCol] != 'R') || (!isWhiteKing && board[0][rookCol] != 'r') {
			return false
		}
		for col := fromCol - 1; col > rookCol; col-- {
			if board[row][col] != 0 || IsSquareUnderAttack(board, row, col, isWhiteKing) {
				return false
			}
		}
	}
	fmt.Println("Castleable")
	return true
}

func IsValidMove(board [8][8]rune, piece rune, fromRow, fromCol, toRow, toCol int, promotionPiece *rune) bool {
	if toRow < 0 || toRow >= 8 || toCol < 0 || toCol >= 8 {
		return false
	}

	if board[toRow][toCol] != 0 {
		if isWhite(piece) == isWhite(board[toRow][toCol]) {
			return false
		}
	}

	switch piece {
	case 'P':
		if fromCol == toCol && board[toRow][toCol] == 0 {
			if toRow == fromRow-1 || (fromRow == 6 && toRow == 4 && board[5][toCol] == 0) {
				return handlePawnPromotion(toRow, promotionPiece, true)
			}
		} else if abs(fromCol-toCol) == 1 && toRow == fromRow-1 && !isWhite(board[toRow][toCol]) {
			return handlePawnPromotion(toRow, promotionPiece, true)
		}
	case 'p':
		if fromCol == toCol && board[toRow][toCol] == 0 {
			if toRow == fromRow+1 || (fromRow == 1 && toRow == 3 && board[2][toCol] == 0) {
				return handlePawnPromotion(toRow, promotionPiece, false)
			}
		} else if abs(fromCol-toCol) == 1 && toRow == fromRow+1 && isWhite(board[toRow][toCol]) {
			return handlePawnPromotion(toRow, promotionPiece, false)
		}
	case 'R', 'r':
		if fromRow == toRow || fromCol == toCol {
			return clearPath(board, fromRow, fromCol, toRow, toCol)
		}
	case 'N', 'n':
		if (abs(fromRow-toRow) == 2 && abs(fromCol-toCol) == 1) || (abs(fromRow-toRow) == 1 && abs(fromCol-toCol) == 2) {
			return true
		}
	case 'B', 'b':
		if abs(fromRow-toRow) == abs(fromCol-toCol) {
			return clearPath(board, fromRow, fromCol, toRow, toCol)
		}
	case 'Q', 'q':
		if fromRow == toRow || fromCol == toCol || abs(fromRow-toRow) == abs(fromCol-toCol) {
			return clearPath(board, fromRow, fromCol, toRow, toCol)
		}
	case 'K', 'k':
		if abs(fromRow-toRow) <= 1 && abs(fromCol-toCol) <= 1 {
			return true
		}
		if IsCastleable(board, fromRow, fromCol, toRow, toCol) {
			return true
		}

	}

	return false
}

func handlePawnPromotion(toRow int, promotionPiece *rune, isWhite bool) bool {
	if (isWhite && toRow == 0) || (!isWhite && toRow == 7) {
		if promotionPiece != nil && (*promotionPiece == 'Q' || *promotionPiece == 'R' || *promotionPiece == 'B' || *promotionPiece == 'N' || *promotionPiece == 'q' || *promotionPiece == 'r' || *promotionPiece == 'b' || *promotionPiece == 'n') {
			fmt.Println("Pawn promotion")
			return true
		}
		fmt.Println("Invalid or missing promotion piece.")
		return false
	}
	return true
}

func clearPath(board [8][8]rune, fromRow, fromCol, toRow, toCol int) bool {
	rowStep := sign(toRow - fromRow)
	colStep := sign(toCol - fromCol)

	row, col := fromRow+rowStep, fromCol+colStep
	for row != toRow || col != toCol {
		if board[row][col] != 0 {
			return false
		}
		row += rowStep
		col += colStep
	}
	return true
}
func sign(x int) int {
	if x < 0 {
		return -1
	} else if x > 0 {
		return 1
	}
	return 0
}

func isWhite(piece rune) bool {
	return piece == 'P' || piece == 'N' || piece == 'B' || piece == 'R' || piece == 'Q' || piece == 'K'
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func GenereateAllMoves(board[8][8] rune,isWhiteTurn bool) []Move{
	var legaMoves []Move
	for fromRow:=0;fromRow<8;fromRow++{
		for fromCol:=0;fromCol<8;fromCol++{
			piece:=board[fromRow][fromCol]
			
			if piece==0 || isWhite(piece) {
				continue
			}
			for toRow:=0;toRow<8;toRow++{
				for toCol:=0;toCol<8;toCol++{
					if IsValidMove(board,piece,fromRow,fromCol,toRow,toCol,nil){
						tempBoard:=board
						tempBoard[toRow][toCol]=piece
						tempBoard[fromRow][fromCol]=0

						kingRow,kingCol:=-1,-1
						for r:=0;r<8;r++{
							for c:=0;c<8;c++{
								if tempBoard[r][c]=='k'{
									kingRow=r
									kingCol=c
									break
								}
							}
						}

						if !IsSquareUnderAttack(tempBoard,kingRow,kingCol,isWhiteTurn){
							newMove:=Move{FromRow:fromRow,FromCol:fromCol,ToRow:toRow,ToCol:toCol}
							legaMoves=append(legaMoves,newMove)
						}
					}
				}
			}
		}
	}

	return legaMoves
}

func FindBestMove(board[8][8] rune,isWhiteTurn bool) Move{
	var bestMove Move
	var bestScore=100000

	allMoves:=GenereateAllMoves(board,isWhiteTurn)

	for _,move := range allMoves {
		tempBoard:=board
		piece:=tempBoard[move.FromRow][move.FromCol]
		tempBoard[move.ToRow][move.ToCol]=piece
		tempBoard[move.FromRow][move.FromCol]=0

		score:=Evaluate_board(tempBoard)

		if score<bestScore{
			bestMove=move
		}
	}
	return bestMove
}