package main

import (
	"chess-engine/handlers"
	"fmt"
	"image/color"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	//"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const boardSize = 8
const pieceDir = "chess-gui/peices"

const startFenNotation = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

var mpPieceToImage = map[rune]string{
	'P': "whitePawn.svg", 'N': "whiteKnight.svg", 'B': "whiteBishop.svg", 'R': "whiteRook.svg",
	'Q': "whiteQueen.svg", 'K': "whiteKing.svg",
	'p': "blackPawn.svg", 'n': "blackKnight.svg", 'b': "blackBishop.svg", 'r': "blackRook.svg",
	'q': "blackQueen.svg", 'k': "blackKing.svg",
}

var selectedRow, selectedCol int
var pieceSelected bool
var parsedBoard [8][8]rune
var whiteTurn = true

var boardContainer *fyne.Container
var boardCells [8][8]*fyne.Container
var blackScore = 1290
var whiteScore = 1290

type KingPosition struct {
	Row     int
	Col     int
	IsCheck bool
}

var whiteKing = KingPosition{Row: 7, Col: 4, IsCheck: false}
var blackKing = KingPosition{Row: 0, Col: 4, IsCheck: false}

func getBlackPieces() [][2]int {
	var positions [][2]int
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece := parsedBoard[i][j]
			if piece >= 'a' && piece <= 'z' {
				positions = append(positions, [2]int{i, j})
			}
		}
	}
	return positions
}

func handlePieceClick(row, col int){
	clickedPiece:=parsedBoard[row][col]

	if !pieceSelected {
		if clickedPiece != 0 && (whiteTurn == isWhite(clickedPiece)) {
			selectedRow, selectedCol = row, col
			pieceSelected = true
			fmt.Printf("Selected piece at: %d, %d\n", row, col)
		}
	} else {

		if clickedPiece != 0 &&
			(whiteTurn == isWhite(clickedPiece)) &&
			(row != selectedRow || col != selectedCol) {
			selectedRow, selectedCol = row, col
			fmt.Printf("Reselected piece at: %d, %d\n", row, col)
			return
		}

		if row == selectedRow && col == selectedCol {
			pieceSelected = false
			fmt.Println("Piece deselected")
			return
		}

		movePiece(selectedRow, selectedCol, row, col)
	}
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

func isPathClear(fromRow, fromCol, toRow, toCol int) bool {
	rowStep, colStep := 0, 0
	piece := parsedBoard[fromRow][fromCol]
	if piece == 'N' || piece == 'n' {
		return true
	}
	if fromRow < toRow {
		rowStep = 1
	} else if fromRow > toRow {
		rowStep = -1
	}

	if fromCol < toCol {
		colStep = 1
	} else if fromCol > toCol {
		colStep = -1
	}

	r, c := fromRow+rowStep, fromCol+colStep
	for r != toRow || c != toCol {
		if parsedBoard[r][c] != 0 {
			return false 
		}
		r += rowStep
		c += colStep
	}

	return true
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func movePiece(fromRow, fromCol, toRow, toCol int) {
	if fromRow == toRow && fromCol == toCol {
		pieceSelected = false
		return
	}

	piece := parsedBoard[fromRow][fromCol]
	isWhitePiece := piece >= 'A' && piece <= 'Z'
	targetPiece := parsedBoard[toRow][toCol]
	isTargetWhitePiece := targetPiece >= 'A' && targetPiece <= 'Z'

	if (whiteTurn && !isWhitePiece) || (!whiteTurn && isWhitePiece) {
		fmt.Println("Not your turn!")
		return
	}

	if (piece == 'K' || piece == 'k') && abs(fromCol-toCol) == 2 {
		if handlers.IsCastleable(parsedBoard, fromRow, fromCol, toRow, toCol) {
			performCastling(fromRow, fromCol, toRow, toCol, piece)
			return
		}
		fmt.Println("Not Possible to castle")
		pieceSelected = false
		return
	}
	if targetPiece != 0 && ((whiteTurn && isTargetWhitePiece) || (!whiteTurn && !isTargetWhitePiece)) {
		fmt.Println("Can't capture own piece")
		return
	}

	if targetPiece != 0 {
		if isWhitePiece {
			blackScore -= handlers.GetValue(targetPiece)
		} else {
			whiteScore -= handlers.GetValue(targetPiece)
		}
	}
	if !isPathClear(fromRow, fromCol, toRow, toCol) {
		fmt.Println("Path is blocked for piece:", string(piece))
		return
	}
	var promotionPiece rune
	if whiteTurn && toRow == 0 && piece == 'P' {
		promotionPiece = 'Q'
	} else if !whiteTurn && toRow == 7 && piece == 'p' {
		promotionPiece = 'q'
	}
	if !handlers.IsValidMove(parsedBoard, piece, fromRow, fromCol, toRow, toCol, &promotionPiece) {
		fmt.Println("Invalid move for piece:", string(piece))
		return
	}
	tempBoard := [8][8]rune{}
	for i:=0;i<8;i++{
		for j := 0; j < 8; j++ {
			tempBoard[i][j] = parsedBoard[i][j]
		}
	}
	tempBoard[toRow][toCol] = piece
	tempBoard[fromRow][fromCol] = 0
	kingToCheck := &whiteKing
	opponentKing := &blackKing
	if !isWhitePiece {
		kingToCheck = &blackKing
		opponentKing = &whiteKing
	}
	if piece == 'K' {
		whiteKing.Row, whiteKing.Col = toRow, toCol
	} else if piece == 'k' {
		blackKing.Row, blackKing.Col = toRow, toCol
	}
	if whiteKing.IsCheck && whiteTurn {
		if !handlers.IsSquareUnderAttack(parsedBoard, whiteKing.Row, whiteKing.Col, true) {
			whiteKing.IsCheck = false
		}
	} else if blackKing.IsCheck && !whiteTurn {
		if !handlers.IsSquareUnderAttack(parsedBoard, blackKing.Row, blackKing.Col, false) {
			blackKing.IsCheck = false
		}
	}
	if handlers.IsSquareUnderAttack(tempBoard, kingToCheck.Row, kingToCheck.Col, isWhitePiece) {
		fmt.Println("Move would leave your king in check!")
		return
	}
	if promotionPiece != 0 {
		piece = promotionPiece
	}
	parsedBoard[toRow][toCol] = piece
	parsedBoard[fromRow][fromCol] = 0

	if piece == 'K' {
		whiteKing.Row, whiteKing.Col = toRow, toCol
	} else if piece == 'k' {
		blackKing.Row, blackKing.Col = toRow, toCol
	}

	if handlers.IsSquareUnderAttack(parsedBoard, opponentKing.Row, opponentKing.Col, !isWhitePiece) {
		if isWhitePiece {
			fmt.Println("Black KING is under check")
			blackKing.IsCheck = true
			// Check for checkmate
			if isCheckmate(false) {
				fmt.Println("Checkmate! White wins!")
			}
		} else {
			fmt.Println("White KING is under check")
			whiteKing.IsCheck = true
			// Check for checkmate
			if isCheckmate(true) {
				fmt.Println("Checkmate! Black wins!")
			}
		}
	} else {
		if !isWhitePiece {
			blackKing.IsCheck = false
		} else {
			whiteKing.IsCheck = false
		}
	}
	fmt.Printf("Moved %c from (%d, %d) to (%d, %d)\n", piece, fromRow, fromCol, toRow, toCol)
	whiteTurn = !whiteTurn
	pieceSelected = false

	updateBoardUI(fromRow, fromCol, toRow, toCol)

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if parsedBoard[i][j] != 0 {
				fmt.Printf("%c ", parsedBoard[i][j])
			} else {
				fmt.Printf("  ")
			}

		}
		fmt.Println()
	}
	//fmt.Println(whiteKing.IsCheck, blackKing.IsCheck)
	fmt.Println("White Score:", whiteScore, "Black Score:", blackScore)

}

func isCheckmate(isWhiteKing bool) bool {
	kingPos := &blackKing
	if isWhiteKing {
		kingPos = &whiteKing
	}

	fmt.Printf("\n=== CHECKMATE ANALYSIS ===\n")
	fmt.Printf("Analyzing position for %s king at (%d,%d)\n",
		map[bool]string{true: "White", false: "Black"}[isWhiteKing],
		kingPos.Row, kingPos.Col)

	// First verify the king is actually in check
	if !kingPos.IsCheck {
		fmt.Println("King is not in check - cannot be checkmate")
		return false
	}

	fmt.Println("\nChecking all possible king moves:")
	// Try all possible king moves
	kingMoves := [][2]int{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}

	kingPiece := 'k'
	if isWhiteKing {
		kingPiece = 'K'
	}

	// Check each possible king move
	for _, move := range kingMoves {
		newRow := kingPos.Row + move[0]
		newCol := kingPos.Col + move[1]

		if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 {
			// Create a deep copy of the board
			tempBoard := [8][8]rune{}
			for i := 0; i < 8; i++ {
				for j := 0; j < 8; j++ {
					tempBoard[i][j] = parsedBoard[i][j]
				}
			}

			targetPiece := tempBoard[newRow][newCol]
			isTargetWhitePiece := targetPiece >= 'A' && targetPiece <= 'Z'

			fmt.Printf("Testing king move to (%d,%d): ", newRow, newCol)

			if targetPiece == 0 || (isWhiteKing != isTargetWhitePiece) {
				tempBoard[newRow][newCol] = kingPiece
				tempBoard[kingPos.Row][kingPos.Col] = 0
				if !handlers.IsSquareUnderAttack(tempBoard, newRow, newCol, isWhiteKing) {
					fmt.Println("LEGAL MOVE FOUND - not checkmate")
					return false
				}
				fmt.Println("square is under attack")

			} else {
				fmt.Println("square occupied by friendly piece")
			}
		}
	}

	fmt.Println("\nCHECKMATE CONFIRMED!")
	return true
}

func performCastling(fromRow, fromCol, toRow, toCol int, piece rune) {
	isKingSide := toCol > fromCol
	rookFromCol := 0
	rookToCol := 3
	if isKingSide {
		rookFromCol = 7
		rookToCol = 5
	}

	parsedBoard[toRow][toCol] = piece
	parsedBoard[fromRow][fromCol] = 0

	rook := 'R'
	if piece == 'k' {
		rook = 'r'
	}
	parsedBoard[toRow][rookToCol] = rook
	parsedBoard[toRow][rookFromCol] = 0

	updateBoardUI(fromRow, fromCol, toRow, toCol)
	updateBoardUI(toRow, rookFromCol, toRow, rookToCol)

	whiteTurn = !whiteTurn
	pieceSelected = false
}

func updateBoardUI(fromRow, fromCol, toRow, toCol int) {

	fromCell := boardCells[fromRow][fromCol]
	toCell := boardCells[toRow][toCol]

	fromSquareColor := color.White
	if (fromRow+fromCol)%2 == 1 {
		fromSquareColor = color.Black
	}
	toSquareColor := color.White
	if (toRow+toCol)%2 == 1 {
		toSquareColor = color.Black
	}

	fromSquare := canvas.NewRectangle(fromSquareColor)
	toSquare := canvas.NewRectangle(toSquareColor)
	fromSquare.SetMinSize(fyne.NewSize(75, 75))
	toSquare.SetMinSize(fyne.NewSize(75, 75))

	fromButton := widget.NewButton(" ", func() {
		handlePieceClick(fromRow, fromCol)
	})
	toButton := widget.NewButton(" ", func() {
		handlePieceClick(toRow, toCol)
	})
	fromButton.Importance = widget.LowImportance
	toButton.Importance = widget.LowImportance
	fromButton.Resize(fyne.NewSize(75, 75))
	toButton.Resize(fyne.NewSize(75, 75))

	fromCell.Objects = []fyne.CanvasObject{fromSquare, fromButton}

	toCell.Objects = []fyne.CanvasObject{toSquare}
	if piece := parsedBoard[toRow][toCol]; piece != 0 {
		imagePath := filepath.Join(pieceDir, mpPieceToImage[piece])
		pieceImage := canvas.NewImageFromFile(imagePath)
		pieceImage.FillMode = canvas.ImageFillContain
		pieceImage.Resize(fyne.NewSize(75, 75))
		toCell.Objects = append(toCell.Objects, pieceImage)
	}
	toCell.Objects = append(toCell.Objects, toButton)

	fromCell.Refresh()
	toCell.Refresh()
}
func isWhite(piece rune) bool {
	return piece == 'P' || piece == 'N' || piece == 'B' || piece == 'R' || piece == 'Q' || piece == 'K'
}

func generateChessBoard() *fyne.Container {
	board := container.NewGridWithColumns(boardSize)

	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			squareColor := color.White
			if (row+col)%2 == 1 {
				squareColor = color.Black
			}
			square := canvas.NewRectangle(squareColor)
			square.SetMinSize(fyne.NewSize(75, 75))

			cell := container.NewStack(square)

			rowCopy, colCopy := row, col
			tapButton := widget.NewButton(" ", func() {
				handlePieceClick(rowCopy, colCopy)
			})
			tapButton.Importance = widget.LowImportance
			tapButton.Resize(fyne.NewSize(75, 75))

			if piece := parsedBoard[row][col]; piece != 0 {
				imagePath := filepath.Join(pieceDir, mpPieceToImage[piece])
				pieceImage := canvas.NewImageFromFile(imagePath)
				pieceImage.FillMode = canvas.ImageFillContain
				pieceImage.Resize(fyne.NewSize(75, 75))
				cell = container.NewStack(square, pieceImage, tapButton)
			} else {
				cell = container.NewStack(square, tapButton)
			}

			boardCells[row][col] = cell
			board.Add(cell)
		}
	}

	return board
}

func main() {
	chessApp := app.New()
	//a := app.New()
	//w := a.NewWindow("Chess Game")
	window := chessApp.NewWindow("Chess Game")
	window.Resize(fyne.NewSize(600, 600))

	parsedBoard = parseFEN(strings.Split(startFenNotation, " ")[0])

	boardContainer = container.NewVBox(
		widget.NewLabel("Chess Game"),
		generateChessBoard(),
	)

	window.SetContent(boardContainer)
	//showWinnerNotification(a, w, "White")
	window.ShowAndRun()
}
