package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBinaryEncoding(t *testing.T) {
	g := NewGame()
	// Make some moves to populate history and complex state
	err := g.LoadPGN("1. e4 e5 2. Nf3 Nc6 3. Bb5 a6")
	require.NoError(t, err)

	data, err := g.MarshalBinary()
	require.NoError(t, err)

	b2 := &Game{}
	err = b2.UnmarshalBinary(data)
	require.NoError(t, err)

	require.Equal(t, g.turn, b2.turn)
	require.Equal(t, g.castling, b2.castling)
	require.Equal(t, g.enPassant, b2.enPassant)
	require.Equal(t, g.halfMoves, b2.halfMoves)
	require.Equal(t, g.fullMoves, b2.fullMoves)
	require.Equal(t, g.zobristHash, b2.zobristHash)
	require.Equal(t, g.Occupation(), b2.Occupation())
	require.Equal(t, len(g.positionHistory), len(b2.positionHistory))

	for k, v := range g.positionHistory {
		require.Equal(t, v, b2.positionHistory[k], "History count mismatch for hash %x", k)
	}

	// Verify move generation is consistent
	g.GenerateLegalMoves()
	b2.GenerateLegalMoves()
	require.Equal(t, len(g.legalMoves), len(b2.legalMoves))
	for i := range g.legalMoves {
		require.Equal(t, g.legalMoves[i], b2.legalMoves[i])
	}
}

func TestMarshalBinaryRoundTripsChess960Variant(t *testing.T) {
	g, err := NewGameFromFENWithVariant(STARTING_POSITION_FEN, VariantChess960)
	require.NoError(t, err)

	data, err := g.MarshalBinary()
	require.NoError(t, err)

	var decoded Game
	require.NoError(t, decoded.UnmarshalBinary(data))
	require.Equal(t, VariantChess960, decoded.Variant())
	require.Equal(t, g.FEN(), decoded.FEN())
	require.Equal(t, g.PositionKey(), decoded.PositionKey())
}

func TestMarshalBinaryRoundTripsThreeCheckVariant(t *testing.T) {
	g, err := NewGameFromFENWithVariant("4k3/8/8/8/8/8/Q7/4K3 w - - 0 1 +2+0", VariantThreeCheck)
	require.NoError(t, err)

	data, err := g.MarshalBinary()
	require.NoError(t, err)

	var decoded Game
	require.NoError(t, decoded.UnmarshalBinary(data))
	require.Equal(t, VariantThreeCheck, decoded.Variant())
	require.Equal(t, g.FEN(), decoded.FEN())
	require.Equal(t, g.PositionKey(), decoded.PositionKey())
}

func TestUnmarshalBinaryRejectsInvalidThreeCheckCounters(t *testing.T) {
	g, err := NewGameFromFENWithVariant("4k3/8/8/8/8/8/Q7/4K3 w - - 0 1 +2+0", VariantThreeCheck)
	require.NoError(t, err)

	data, err := g.MarshalBinary()
	require.NoError(t, err)

	variantOffset := binaryHeaderSize + 64 + 15
	data[variantOffset+1] = 4

	var decoded Game
	require.Error(t, decoded.UnmarshalBinary(data))
}

func TestMarshalBinaryRoundTripsKingOfTheHillVariant(t *testing.T) {
	g, err := NewGameFromFENWithVariant("4k3/8/8/8/8/3K4/8/R7 w - - 0 1", VariantKingOfTheHill)
	require.NoError(t, err)

	data, err := g.MarshalBinary()
	require.NoError(t, err)

	var decoded Game
	require.NoError(t, decoded.UnmarshalBinary(data))
	require.Equal(t, VariantKingOfTheHill, decoded.Variant())
	require.Equal(t, g.FEN(), decoded.FEN())
	require.Equal(t, g.PositionKey(), decoded.PositionKey())
}

func TestBinaryDecodingRejectsInvalidPayloads(t *testing.T) {
	g := NewGame()
	data, err := g.MarshalBinary()
	require.NoError(t, err)

	t.Run("bad magic", func(t *testing.T) {
		bad := append([]byte(nil), data...)
		bad[0] = 'X'

		decoded := &Game{}
		require.Error(t, decoded.UnmarshalBinary(bad))
	})

	t.Run("bad piece", func(t *testing.T) {
		bad := append([]byte(nil), data...)
		bad[4] = 255

		decoded := &Game{}
		require.Error(t, decoded.UnmarshalBinary(bad))
	})

	t.Run("truncated", func(t *testing.T) {
		decoded := &Game{}
		require.Error(t, decoded.UnmarshalBinary(data[:len(data)-1]))
	})
}

func (g *Game) Occupation() uint64 {
	return uint64(g.occupied)
}
