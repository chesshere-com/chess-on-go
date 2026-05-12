package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSEE_EmptyFromReturnsZero(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("4k3/8/8/8/8/8/8/4K3 w - - 0 1"))
	from, err := ParseSquare("e4")
	require.NoError(t, err)
	to, err := ParseSquare("d5")
	require.NoError(t, err)
	require.Equal(t, 0, g.SEE(from, to))
}

func TestSEE_OutOfRangeReturnsZero(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("4k3/8/8/8/8/8/8/4K3 w - - 0 1"))
	require.Equal(t, 0, g.SEE(Square(64), Square(0)))
	require.Equal(t, 0, g.SEE(Square(0), Square(64)))
	require.Equal(t, 0, g.SEE(Square(64), Square(64)))
}
