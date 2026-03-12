package handlers

import (
	"fmt"
	"sort"
	"time"
)

type CastlingRights struct {
	WhiteKingSide  bool
	WhiteQueenSide bool
	BlackKingSide  bool
	BlackQueenSide bool
}

// type HashMap struct {
// 	Hash uint64
// 	Score int
// 	Depth int
// 	BestMove Move
// }

type Move struct {
	FromRow, FromCol int
	ToRow, ToCol     int
}

// Simple in-engine profiling counters (aggregated across calls).
var (
	IsValidMoveTime         time.Duration
	IsValidMoveCount        int64
	GenerateAllMovesTime    time.Duration
	GenerateAllMovesCount   int64
	GenerateCaptureMovesTime time.Duration
	GenerateCaptureMovesCount int64
	FindBestMoveTime        time.Duration
	FindBestMoveCount       int64
	MinimaxTime             time.Duration
	MinimaxCount            int64
	QuiescenceTime          time.Duration
	QuiescenceCount         int64
)

// ResetProfiling clears all profiling counters; useful between moves.
func ResetProfiling() {
	IsValidMoveTime = 0
	IsValidMoveCount = 0
	GenerateAllMovesTime = 0
	GenerateAllMovesCount = 0
	GenerateCaptureMovesTime = 0
	GenerateCaptureMovesCount = 0
	FindBestMoveTime = 0
	FindBestMoveCount = 0
	MinimaxTime = 0
	MinimaxCount = 0
	QuiescenceTime = 0
	QuiescenceCount = 0
}

// var initialPositions = map[string]bool{
// 	"e1": true,
// 	"e8": true,
// 	"a1": true,
// 	"h1": true,
// 	"a8": true,
// 	"h8": true,
// }

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

func IsSquareUnderAttack(board [8][8]rune, row, col int, attackerIsWhite bool) bool {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece := board[i][j]
			if piece == 0 || isWhite(piece) != attackerIsWhite {
				continue
			}
			if canAttackSquare(board, piece, i, j, row, col) {
				return true
			}
		}
	}
	return false
}

func canAttackSquare(board [8][8]rune, piece rune, fromRow, fromCol, toRow, toCol int) bool {
	// Bounds check to prevent index out of range errors
	if toRow < 0 || toRow >= 8 || toCol < 0 || toCol >= 8 {
		return false
	}
	if fromRow < 0 || fromRow >= 8 || fromCol < 0 || fromCol >= 8 {
		return false
	}
	if board[toRow][toCol] != 0 && isWhite(piece) == isWhite(board[toRow][toCol]) {
		return false
	}

	switch piece {
	case 'P':
		return abs(fromCol-toCol) == 1 && toRow == fromRow-1
	case 'p':
		return abs(fromCol-toCol) == 1 && toRow == fromRow+1
	case 'R', 'r':
		return (fromRow == toRow || fromCol == toCol) && clearPath(board, fromRow, fromCol, toRow, toCol)
	case 'N', 'n':
		return (abs(fromRow-toRow) == 2 && abs(fromCol-toCol) == 1) ||
			(abs(fromRow-toRow) == 1 && abs(fromCol-toCol) == 2)
	case 'B', 'b':
		return abs(fromRow-toRow) == abs(fromCol-toCol) && clearPath(board, fromRow, fromCol, toRow, toCol)
	case 'Q', 'q':
		return (fromRow == toRow || fromCol == toCol || abs(fromRow-toRow) == abs(fromCol-toCol)) &&
			clearPath(board, fromRow, fromCol, toRow, toCol)
	case 'K', 'k':
		return abs(fromRow-toRow) <= 1 && abs(fromCol-toCol) <= 1
	}
	return false
}

func IsInCheck(board [8][8]rune, isWhiteKing bool, kingRow, kingCol int) bool {
	// If king is not found, return false (shouldn't happen in valid game states)
	if kingRow < 0 || kingRow >= 8 || kingCol < 0 || kingCol >= 8 {
		return false
	}
	return IsSquareUnderAttack(board, kingRow, kingCol, !isWhiteKing)
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
		//fmt.Println("King cant be castled when under check ")
		return false
	}

	if isKingSide {
		rookCol := 7
		if (isWhiteKing && board[7][rookCol] != 'R') || (!isWhiteKing && board[0][rookCol] != 'r') {
			return false
		}
		for col := fromCol + 1; col < rookCol; col++ {
			if board[row][col] != 0 || IsSquareUnderAttack(board, row, col, !isWhiteKing) {
				return false
			}
		}
	} else {
		rookCol := 0
		if (isWhiteKing && board[7][rookCol] != 'R') || (!isWhiteKing && board[0][rookCol] != 'r') {
			return false
		}
		for col := fromCol - 1; col > rookCol; col-- {
			if board[row][col] != 0 || IsSquareUnderAttack(board, row, col, !isWhiteKing) {
				return false
			}
		}
	}
	//fmt.Println("Castleable")
	return true
}

func IsValidMove(board [8][8]rune, piece rune, fromRow, fromCol, toRow, toCol int, promotionPiece *rune) bool {
	start := time.Now()
	defer func() {
		IsValidMoveTime += time.Since(start)
		IsValidMoveCount++
	}()

	if toRow < 0 || toRow >= 8 || toCol < 0 || toCol >= 8 {
		return false
	}

	if board[toRow][toCol] != 0 {
		if isWhite(piece) == isWhite(board[toRow][toCol]) {
			return false
		}
	}
	validPieceMove := false
	switch piece {
	case 'P':
		if fromCol == toCol && board[toRow][toCol] == 0 {
			if toRow == fromRow-1 || (fromRow == 6 && toRow == 4 && board[5][toCol] == 0) {
				validPieceMove = handlePawnPromotion(toRow, 'Q', true)
			}
		} else if abs(fromCol-toCol) == 1 && toRow == fromRow-1 && !isWhite(board[toRow][toCol]) && board[toRow][toCol] != 0 {
			validPieceMove = handlePawnPromotion(toRow, 'Q', true)
		}
	case 'p':
		if fromCol == toCol && board[toRow][toCol] == 0 {
			if toRow == fromRow+1 || (fromRow == 1 && toRow == 3 && board[2][toCol] == 0) {
				validPieceMove = handlePawnPromotion(toRow, 'q', false)
			}
		} else if abs(fromCol-toCol) == 1 && toRow == fromRow+1 && isWhite(board[toRow][toCol]) {
			validPieceMove = handlePawnPromotion(toRow, 'q', false)
		}
	case 'R', 'r':
		if fromRow == toRow || fromCol == toCol {
			validPieceMove = clearPath(board, fromRow, fromCol, toRow, toCol)
		}
	case 'N', 'n':
		if (abs(fromRow-toRow) == 2 && abs(fromCol-toCol) == 1) || (abs(fromRow-toRow) == 1 && abs(fromCol-toCol) == 2) {
			validPieceMove = true
		}
	case 'B', 'b':
		if abs(fromRow-toRow) == abs(fromCol-toCol) {
			validPieceMove = clearPath(board, fromRow, fromCol, toRow, toCol)
		}
	case 'Q', 'q':
		if fromRow == toRow || fromCol == toCol || abs(fromRow-toRow) == abs(fromCol-toCol) {
			validPieceMove = clearPath(board, fromRow, fromCol, toRow, toCol)
		}
	case 'K', 'k':
		if IsSquareUnderAttack(board, toRow, toCol, !isWhite(piece)) {
			return false
		}
		if abs(fromRow-toRow) <= 1 && abs(fromCol-toCol) <= 1 {
			validPieceMove = true
		} else if abs(fromCol-toCol) == 2 && fromRow == toRow {
			validPieceMove = IsCastleable(board, fromRow, fromCol, toRow, toCol)
		}
	}
	if !validPieceMove {
		return false
	}
	kingRow, kingCol := -1, -1
	kingPiece := 'k'
	if isWhite(piece) {
		kingPiece = 'K'
	}

	var tempBoard [8][8]rune
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			tempBoard[r][c] = board[r][c]
			if board[r][c] == kingPiece {
				kingRow = r
				kingCol = c
			}
		}
	}

	tempBoard[fromRow][fromCol] = 0
	tempBoard[toRow][toCol] = piece

	if piece == 'K' || piece == 'k' {
		kingRow = toRow
		kingCol = toCol
	}

	if kingRow == -1 || kingCol == -1 {
		return false
	}

	if IsInCheck(tempBoard, isWhite(piece), kingRow, kingCol) {
		return false
	}

	return true
}
func handlePawnPromotion(toRow int, promotionPiece rune, isWhite bool) bool {
	if (isWhite && toRow == 0) || (!isWhite && toRow == 7) {
		if promotionPiece == 'Q' || promotionPiece == 'R' || promotionPiece == 'B' || promotionPiece == 'N' || promotionPiece == 'q' || promotionPiece == 'r' || promotionPiece == 'b' || promotionPiece == 'n' {
			//fmt.Println("Pawn promotion")
			return true
		}
		//fmt.Println("Invalid or missing promotion piece.")
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

func score_move(move Move, board [8][8]rune) int {
	startRow := move.FromRow
	startCol := move.FromCol
	endRow := move.ToRow
	endCol := move.ToCol

	current_piece := board[startRow][startCol]
	next_piece := board[endRow][endCol]

	score := 0

	if next_piece != 0 {
		score = abs(GetValue(next_piece)) - abs(GetValue(current_piece)/10)
		score = score + 10000
	}
	var tempBoard [8][8]rune
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			tempBoard[i][j] = board[i][j]
		}
	}
	prev := Evaluate_board(tempBoard)
	tempBoard[endRow][endCol] = current_piece
	tempBoard[startRow][startCol] = 0
	after := Evaluate_board(tempBoard)
	//black promition and position changes
	if !isWhite(current_piece) {
		if startRow == 6 && endRow == 7 {
			score += 800
		}
		if after-prev < 0 {
			score += abs(after - prev)
		} else {
			score -= abs(after - prev)
		}
	}
	//white promotion and psoition changes
	if isWhite(current_piece) {
		if startRow == 1 && endRow == 0 {
			score += 800
		}
		score += after - prev
	}
	return score

}

func findKing(board [8][8]rune, isWhite bool) (int, int) {
	kingToFind:='K'
	if !isWhite {
		kingToFind='k'
	}
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if board[i][j] == kingToFind {
				return i, j
			}
		}
	}	
	return -1, -1
}

func GenereateAllMoves(board [8][8]rune, isWhiteTurn bool) []Move {
	start := time.Now()
	defer func() {
		GenerateAllMovesTime += time.Since(start)
		GenerateAllMovesCount++
	}()

	var legalMoves []Move
	for fromRow := 0; fromRow < 8; fromRow++ {
		for fromCol := 0; fromCol < 8; fromCol++ {
			piece := board[fromRow][fromCol]

			if piece == 0 {
				continue
			}
			if isWhiteTurn && !isWhite(piece) {
				continue
			}
			if !isWhiteTurn && isWhite(piece) {
				continue
			}

			possibleMoves := getPossibleMoves(piece, fromRow, fromCol,board)
			for _,pos:=range possibleMoves{
				// Bounds check to prevent index out of range errors
				if pos[0] < 0 || pos[0] >= 8 || pos[1] < 0 || pos[1] >= 8 {
					continue
				}
				var tempBoard [8][8]rune
				for i:=0;i<8;i++{
					for j:=0;j<8;j++{
						tempBoard[i][j]=board[i][j]
					}
				}
				tempBoard[pos[0]][pos[1]]=piece
				tempBoard[fromRow][fromCol]=0				
	
				var kingRow, kingCol int
				if piece == 'K' || piece == 'k' {
					kingRow = pos[0]
					kingCol = pos[1]
				} else {
					kingRow, kingCol = findKing(board, isWhiteTurn)
				}
				
				if !IsInCheck(tempBoard, isWhiteTurn, kingRow, kingCol){
					legalMoves = append(legalMoves, Move{FromRow: fromRow, FromCol: fromCol, ToRow: pos[0], ToCol: pos[1]})
				}
			}
			// for _, pos := range possibleMoves {
			// 	toRow, toCol := pos[0], pos[1]
			// 	if IsValidMove(board, piece, fromRow, fromCol, toRow, toCol, nil) {
			// 		legalMoves = append(legalMoves, Move{FromRow: fromRow, FromCol: fromCol, ToRow: toRow, ToCol: toCol})
			// 	}
			// }
			// for toRow := 0; toRow < 8; toRow++ {
			// 	for toCol := 0; toCol < 8; toCol++ {
			// 		if IsValidMove(board, piece, fromRow, fromCol, toRow, toCol, nil) {
			// 			newMove := Move{FromRow: fromRow, FromCol: fromCol, ToRow: toRow, ToCol: toCol}
			// 			legalMoves = append(legalMoves, newMove)
			// 		}
			// 	}
			// }
		}
	}

	sort.Slice(legalMoves, func(i, j int) bool {
		score_i := score_move(legalMoves[i], board)
		score_j := score_move(legalMoves[j], board)
		return score_i > score_j
	})

	return legalMoves
}

func getPossibleMoves(piece rune, fromRow, fromCol int, board [8][8]rune) [][2]int {
	var moves [][2]int

	switch piece {
	case 'N', 'n':
		deltas := [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}
		for _, d := range deltas {
			r, c := fromRow+d[0], fromCol+d[1]
			if r >= 0 && r < 8 && c >= 0 && c < 8 {
				if board[r][c] == 0 || isWhite(piece) != isWhite(board[r][c]) {
					moves = append(moves, [2]int{r, c})
				}
			}
		}

	case 'K', 'k':
		// Normal king moves (one square in any direction)
		for dr := -1; dr <= 1; dr++ {
			for dc := -1; dc <= 1; dc++ {
				if dr == 0 && dc == 0 {
					continue
				}

				r, c := fromRow+dr, fromCol+dc
				if r >= 0 && r < 8 && c >= 0 && c < 8 {
					if board[r][c] == 0 || isWhite(piece) != isWhite(board[r][c]) {
						moves = append(moves, [2]int{r, c})
					}
				}
			}
		}
		if fromCol+2 < 8 && IsCastleable(board, fromRow, fromCol, fromRow, fromCol+2) {
			moves = append(moves, [2]int{fromRow, fromCol + 2})
		}
		// Queen-side
		if fromCol-2 >= 0 && IsCastleable(board, fromRow, fromCol, fromRow, fromCol-2) {
			moves = append(moves, [2]int{fromRow, fromCol - 2})
		}

	case 'P':
		if fromRow > 0 && board[fromRow - 1][fromCol] == 0 {
			moves = append(moves, [2]int{fromRow - 1, fromCol})
		}
		if fromRow == 6 && board[4][fromCol] == 0 && board[5][fromCol] == 0 {
			moves = append(moves, [2]int{4, fromCol})
		}
		if fromRow > 0 && fromCol > 0 && isWhite(piece) != isWhite(board[fromRow - 1][fromCol - 1]) {
			moves = append(moves, [2]int{fromRow - 1, fromCol - 1})
		}
		if fromRow > 0 && fromCol < 7 && isWhite(piece) != isWhite(board[fromRow - 1][fromCol + 1]) {
			moves = append(moves, [2]int{fromRow - 1, fromCol + 1})
		}
	case 'p':
		if fromRow < 7 && board[fromRow + 1][fromCol] == 0 {
			moves = append(moves, [2]int{fromRow + 1, fromCol})
		}
		if fromRow == 1 && board[3][fromCol] == 0 && board[2][fromCol] == 0 {
			moves = append(moves, [2]int{3, fromCol})
		}
		if fromRow < 7 && fromCol > 0 && isWhite(piece) != isWhite(board[fromRow + 1][fromCol - 1]) {
			moves = append(moves, [2]int{fromRow + 1, fromCol - 1})
		}
		if fromRow < 7 && fromCol < 7 && isWhite(piece) != isWhite(board[fromRow + 1][fromCol + 1]) {
			moves = append(moves, [2]int{fromRow + 1, fromCol + 1})
		}
	case 'R', 'r':
		directions:=[][2]int{{-1,0},{1,0},{0,-1},{0,1}}
		for _,d := range directions {
			for i:=1;i<8;i++{
				toRow,toCol:=fromRow+d[0]*i,fromCol+d[1]*i
				if toRow<0 || toRow>=8 || toCol<0 || toCol>=8 {
					break
				}
				if board[toRow][toCol]!=0{
					if isWhite(piece) != isWhite(board[toRow][toCol]) {
						moves = append(moves, [2]int{toRow, toCol})
					}
					break
				}
				moves = append(moves, [2]int{toRow, toCol})
			}
		}
	case 'B', 'b':
		directions:=[][2]int{{-1,-1},{-1,1},{1,-1},{1,1}}
		for _,d := range directions {
			for i:=1;i<8;i++{
				toRow,toCol:=fromRow+d[0]*i,fromCol+d[1]*i
				if toRow<0 || toRow>=8 || toCol<0 || toCol>=8 {
					break
				}
				if board[toRow][toCol]!=0{
					if isWhite(piece) != isWhite(board[toRow][toCol]) {
						moves = append(moves, [2]int{toRow, toCol})
					}
					break
				}
				moves = append(moves, [2]int{toRow, toCol})
			}
		}
	case 'Q', 'q':
		directions:=[][2]int{{-1,0},{1,0},{0,-1},{0,1},{-1,-1},{-1,1},{1,-1},{1,1}}
		for _,d := range directions {
			for i:=1;i<8;i++{
				toRow,toCol:=fromRow+d[0]*i,fromCol+d[1]*i
				if toRow<0 || toRow>=8 || toCol<0 || toCol>=8 {
					break
				}
				if board[toRow][toCol]!=0{
					if isWhite(piece) != isWhite(board[toRow][toCol]) {
						moves = append(moves, [2]int{toRow, toCol})
					}
					break
				}
				moves = append(moves, [2]int{toRow, toCol})
			}
		}
	}
	return moves
}

func GenerateCaptureMoves(board [8][8]rune, isWhiteTurn bool) []Move {
	start := time.Now()
	defer func() {
		GenerateCaptureMovesTime += time.Since(start)
		GenerateCaptureMovesCount++
	}()

	allMoves := GenereateAllMoves(board, isWhiteTurn)
	var capturemoves []Move
	for _, move := range allMoves {
		isCapture := board[move.ToRow][move.ToCol] != 0
		piece := board[move.FromRow][move.FromCol]
		isPawn := piece == 'p' || piece == 'P'
		isPromotion := isPawn && (move.ToRow == 0 || move.ToRow == 7)
		if isPromotion || isCapture {
			capturemoves = append(capturemoves, move)
		}
	}
	return capturemoves
}

func FindBestMove(board [8][8]rune, isWhiteTurn bool) Move {
	start := time.Now()
	defer func() {
		FindBestMoveTime += time.Since(start)
		FindBestMoveCount++
	}()

	allMoves := GenereateAllMoves(board, isWhiteTurn)
	if len(allMoves) == 0 {
		fmt.Println("U have lost MINIMAX")
		return Move{}
	}

	initial_hash := GetZobristValue(board)
	index := initial_hash & (ttSize - 1)
	entry := &transpositionTable[index]
	if entry.HashKey == initial_hash && entry.Depth >= 3 {
		//fmt.Println("hash found in the database using it ")
		return transpositionTable[index].BestMove
	}

	// Aspiration Search with Iterative Deepening
	const targetDepth = 3 
	const aspirationWindow = 25 
	const infinity = 100000
	const negInfinity = -100000

	var bestMove Move = allMoves[0]
	var bestScore int
	var previousScore int = 0 
	
	for depth := 1; depth <= targetDepth; depth++ {
		var alpha, beta int
		var score int
		
		if depth > 1 {
			alpha = previousScore - aspirationWindow
			beta = previousScore + aspirationWindow
			
			score, bestMove = searchWithAspiration(board, isWhiteTurn, depth, alpha, beta, initial_hash, allMoves, previousScore)
			
			if score <= alpha {
				alpha = negInfinity
				beta = previousScore + aspirationWindow
				score, bestMove = searchWithAspiration(board, isWhiteTurn, depth, alpha, beta, initial_hash, allMoves, previousScore)
			} else if score >= beta {
				alpha = previousScore - aspirationWindow
				beta = infinity
				score, bestMove = searchWithAspiration(board, isWhiteTurn, depth, alpha, beta, initial_hash, allMoves, previousScore)
			}
		} else {
			alpha = negInfinity
			beta = infinity
			score, bestMove = searchWithAspiration(board, isWhiteTurn, depth, alpha, beta, initial_hash, allMoves, 0)
		}
		
		bestScore = score
		previousScore = score
		
		learnedInfo := HashMap{
			HashKey:  initial_hash,
			Score:    bestScore,
			Depth:    depth,
			BestMove: bestMove,
		}
		transpositionTable[index] = learnedInfo
	}

	return bestMove
}

func searchWithAspiration(board [8][8]rune, isWhiteTurn bool, depth int, alpha, beta int, initial_hash uint64, allMoves []Move, previousScore int) (int, Move) {
	const infinity = 100000
	const negInfinity = -100000
	
	var bestMove Move = allMoves[0]
	var bestScore int
	
	if isWhiteTurn {
		bestScore = negInfinity
	} else {
		bestScore = infinity
	}

	for _, move := range allMoves {
		var tempBoard [8][8]rune
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				tempBoard[i][j] = board[i][j]
			}
		}
		piece := tempBoard[move.FromRow][move.FromCol]
		tempBoard[move.ToRow][move.ToCol] = piece
		tempBoard[move.FromRow][move.FromCol] = 0
		new_hash := UpdateHashForMove(initial_hash, move, board)
		
		score := Minimax(tempBoard, depth, !isWhiteTurn, alpha, beta, new_hash)

		if isWhiteTurn {
			if score > bestScore {
				bestScore = score
				bestMove = move
			}
			if score > alpha {
				alpha = score
			}
			if alpha >= beta {
				break // Beta cutoff
			}
		} else {
			if score < bestScore {
				bestScore = score
				bestMove = move
			}
			if score < beta {
				beta = score
			}
			if alpha >= beta {
				break // Alpha cutoff
			}
		}
	}

	return bestScore, bestMove
}


func SearchSpecificMoves(board [8][8]rune, isWhiteTurn bool, movesToSearch []Move) (Move, int) {
	if len(movesToSearch) == 0 {
		return Move{}, 0
	}

	const targetDepth = 3
	const aspirationWindow = 25
	const infinity = 100000
	const negInfinity = -100000

	initial_hash := GetZobristValue(board)
	var bestMove Move = movesToSearch[0]
	var bestScore int
	var previousScore int = 0

	// Iterative deepening for this subset
	for depth := 1; depth <= targetDepth; depth++ {
		var alpha, beta int
		var score int

		if depth > 1 {
			alpha = previousScore - aspirationWindow
			beta = previousScore + aspirationWindow

			score, bestMove = searchMovesSubset(board, isWhiteTurn, depth, alpha, beta, initial_hash, movesToSearch, previousScore)

			if score <= alpha {
				alpha = negInfinity
				beta = previousScore + aspirationWindow
				score, bestMove = searchMovesSubset(board, isWhiteTurn, depth, alpha, beta, initial_hash, movesToSearch, previousScore)
			} else if score >= beta {
				alpha = previousScore - aspirationWindow
				beta = infinity
				score, bestMove = searchMovesSubset(board, isWhiteTurn, depth, alpha, beta, initial_hash, movesToSearch, previousScore)
			}
		} else {
			alpha = negInfinity
			beta = infinity
			score, bestMove = searchMovesSubset(board, isWhiteTurn, depth, alpha, beta, initial_hash, movesToSearch, 0)
		}

		bestScore = score
		previousScore = score
	}

	return bestMove, bestScore
}

// searchMovesSubset searches only the provided moves subset
func searchMovesSubset(board [8][8]rune, isWhiteTurn bool, depth int, alpha, beta int, initial_hash uint64, movesToSearch []Move, previousScore int) (int, Move) {
	const infinity = 100000
	const negInfinity = -100000

	var bestMove Move = movesToSearch[0]
	var bestScore int

	if isWhiteTurn {
		bestScore = negInfinity
	} else {
		bestScore = infinity
	}

	for _, move := range movesToSearch {
		var tempBoard [8][8]rune
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				tempBoard[i][j] = board[i][j]
			}
		}
		piece := tempBoard[move.FromRow][move.FromCol]
		tempBoard[move.ToRow][move.ToCol] = piece
		tempBoard[move.FromRow][move.FromCol] = 0
		new_hash := UpdateHashForMove(initial_hash, move, board)

		score := Minimax(tempBoard, depth, !isWhiteTurn, alpha, beta, new_hash)

		if isWhiteTurn {
			if score > bestScore {
				bestScore = score
				bestMove = move
			}
			if score > alpha {
				alpha = score
			}
			if alpha >= beta {
				break // Beta cutoff
			}
		} else {
			if score < bestScore {
				bestScore = score
				bestMove = move
			}
			if score < beta {
				beta = score
			}
			if alpha >= beta {
				break // Alpha cutoff
			}
		}
	}

	return bestScore, bestMove
}

func QuiescenceSearch(board [8][8]rune, isWhiteTurn bool, alpha, beta int) int {
	start := time.Now()
	defer func() {
		QuiescenceTime += time.Since(start)
		QuiescenceCount++
	}()
	base_score := Evaluate_board(board)
	if isWhiteTurn {
		if base_score >= beta {
			return beta
		}
		if base_score > alpha {
			alpha = base_score
		}
	} else {
		if base_score <= alpha {
			return alpha
		}
		if base_score < beta {
			beta = base_score
		}
	}
	capture_move := GenerateCaptureMoves(board, isWhiteTurn)
	sort.Slice(capture_move, func(i,j int) bool {
		score_i:=score_move(capture_move[i],board)
		score_j:=score_move(capture_move[j],board)
		return score_i>score_j
	})

	for _, move := range capture_move {
		var tempBoard [8][8]rune
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				tempBoard[i][j] = board[i][j]
			}
		}

		piece := tempBoard[move.FromRow][move.FromCol]
		tempBoard[move.ToRow][move.ToCol] = piece
		tempBoard[move.FromRow][move.FromCol] = 0
		score := QuiescenceSearch(tempBoard, !isWhiteTurn, alpha, beta)
		if isWhiteTurn {
			if score > alpha {
				alpha = score
			} else {
				if score < beta {
					beta = score
				}
			}
			if alpha >= beta {
				break
			}
		}
	}

	if isWhiteTurn {
		return alpha
	} else {
		return beta
	}
}

func Minimax(board [8][8]rune, depth int, isWhiteTurn bool, alpha int, beta int, current_hash uint64) int {
	start := time.Now()
	defer func() {
		MinimaxTime += time.Since(start)
		MinimaxCount++
	}()

	index := current_hash & (ttSize - 1)
	entry := &transpositionTable[index]

	if entry.HashKey == current_hash && entry.Depth >= depth {
		return entry.Score
	}

	if depth == 0 {
		return QuiescenceSearch(board, isWhiteTurn, alpha, beta)
	}

	allMoves := GenereateAllMoves(board, isWhiteTurn)
	if len(allMoves) == 0 {
		return -99999
	}

	var bestMove Move
	var bestScore int

	if isWhiteTurn {
		bestScore = -100000
		for _, move := range allMoves {
			var tempBoard [8][8]rune
			for i := 0; i < 8; i++ {
				for j := 0; j < 8; j++ {
					tempBoard[i][j] = board[i][j]
				}
			}

			new_hash := UpdateHashForMove(current_hash, move, board)
			makeMove(&tempBoard, move)

			score := Minimax(tempBoard, depth-1, !isWhiteTurn, alpha, beta, new_hash)

			if score > bestScore {
				bestScore = score
				bestMove = move
			}
			if score > alpha {
				alpha = score
			}
			if alpha >= beta {
				break
			}
		}
	} else {
		bestScore = 100000
		for _, move := range allMoves {
			var tempBoard [8][8]rune
			for i := 0; i < 8; i++ {
				for j := 0; j < 8; j++ {
					tempBoard[i][j] = board[i][j]
				}
			}

			new_hash := UpdateHashForMove(current_hash, move, board)
			makeMove(&tempBoard, move)

			score := Minimax(tempBoard, depth-1, !isWhiteTurn, alpha, beta, new_hash)

			if score < bestScore {
				bestScore = score
				bestMove = move
			}
			if score < beta {
				beta = score
			}
			if alpha >= beta {
				break
			}
		}
	}

	entry.HashKey = current_hash
	entry.Score = bestScore
	entry.Depth = depth
	entry.BestMove = bestMove

	return bestScore
}

// func max(a, b int) int {
// 	if a > b {
// 		return a
// 	}
// 	return b
// }

// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }

func makeMove(tempBoard *[8][8]rune, move Move) {
	piece := (*tempBoard)[move.FromRow][move.FromCol]
	(*tempBoard)[move.ToRow][move.ToCol] = piece
	(*tempBoard)[move.FromRow][move.FromCol] = 0
}
