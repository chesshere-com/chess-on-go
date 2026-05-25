package chessongo

import (
	"errors"
	"strings"
)

var errInvalidChess960Position = errors.New("invalid Chess960 position")

// Chess960BackRank returns the white back rank for a Chess960 position ID.
func Chess960BackRank(position int) (string, error) {
	if position < 0 || position > 959 {
		return "", errInvalidChess960Position
	}

	var rank [8]byte
	n := position

	lightBishop := (n%4)*2 + 1
	rank[lightBishop] = 'B'
	n /= 4

	darkBishop := (n % 4) * 2
	rank[darkBishop] = 'B'
	n /= 4

	queenIndex := n % 6
	rank[emptySquares(rank)[queenIndex]] = 'Q'
	n /= 6

	knightIndex := n % 10
	knights := knightSquares(emptySquares(rank), knightIndex)
	rank[knights[0]] = 'N'
	rank[knights[1]] = 'N'

	remaining := emptySquares(rank)
	rank[remaining[0]] = 'R'
	rank[remaining[1]] = 'K'
	rank[remaining[2]] = 'R'

	return string(rank[:]), nil
}

// ValidateChess960BackRank verifies that rank is a legal white Chess960 back rank.
func ValidateChess960BackRank(rank string) error {
	if len(rank) != 8 {
		return errInvalidChess960Position
	}

	counts := map[byte]int{
		'B': 0,
		'K': 0,
		'N': 0,
		'Q': 0,
		'R': 0,
	}
	bishopParity := -1
	kingIndex := -1
	rookBeforeKing := false
	rookAfterKing := false

	for i := 0; i < len(rank); i++ {
		piece := rank[i]
		if _, ok := counts[piece]; !ok {
			return errInvalidChess960Position
		}
		counts[piece]++

		switch piece {
		case 'B':
			if bishopParity == -1 {
				bishopParity = i % 2
			} else if bishopParity == i%2 {
				return errInvalidChess960Position
			}
		case 'K':
			kingIndex = i
		}
	}

	if counts['B'] != 2 || counts['K'] != 1 || counts['N'] != 2 || counts['Q'] != 1 || counts['R'] != 2 {
		return errInvalidChess960Position
	}

	for i := 0; i < len(rank); i++ {
		if rank[i] != 'R' {
			continue
		}
		if i < kingIndex {
			rookBeforeKing = true
		}
		if i > kingIndex {
			rookAfterKing = true
		}
	}
	if !rookBeforeKing || !rookAfterKing {
		return errInvalidChess960Position
	}

	return nil
}

// Chess960StartingFEN returns the initial Shredder-FEN position for a Chess960 position ID.
func Chess960StartingFEN(position int) (string, error) {
	whiteRank, err := Chess960BackRank(position)
	if err != nil {
		return "", err
	}

	blackRank := strings.ToLower(whiteRank)
	return blackRank + "/pppppppp/8/8/8/8/PPPPPPPP/" + whiteRank + " w " + chess960InitialCastlingField(whiteRank) + " - 0 1", nil
}

func chess960InitialCastlingField(whiteRank string) string {
	var files []byte
	for i := 0; i < len(whiteRank); i++ {
		if whiteRank[i] == 'R' {
			files = append(files, byte('A'+i))
		}
	}

	return string([]byte{files[1], files[0], files[1] + ('a' - 'A'), files[0] + ('a' - 'A')})
}

func emptySquares(rank [8]byte) []int {
	var squares []int
	for i, piece := range rank {
		if piece == 0 {
			squares = append(squares, i)
		}
	}
	return squares
}

func knightSquares(squares []int, index int) [2]int {
	current := 0
	for first := 0; first < len(squares)-1; first++ {
		for second := first + 1; second < len(squares); second++ {
			if current == index {
				return [2]int{squares[first], squares[second]}
			}
			current++
		}
	}
	return [2]int{}
}
