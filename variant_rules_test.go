package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVariantRulesLookup(t *testing.T) {
	require.Equal(t, "standard", rulesForVariant(VariantStandard).name)
	require.Equal(t, VariantStandard, rulesForVariant(VariantStandard).variant)
	require.Equal(t, "unknown", rulesForVariant(Variant(99)).name)
}

func TestStandardVariantRulesAreNoOps(t *testing.T) {
	g := NewGame()
	rules := g.rules()

	require.Equal(t, VariantStandard, rules.variant)
	require.Nil(t, rules.afterMove)
	require.Nil(t, rules.overrideStatus)
	require.Nil(t, rules.hashExtra)
}

func TestVariantStateCloneIsIndependent(t *testing.T) {
	g := NewGame()
	g.variant = VariantThreeCheck
	g.variantState.checksGiven[whiteStateIndex] = 2

	clone := g.Clone()
	clone.variantState.checksGiven[whiteStateIndex] = 1

	require.Equal(t, uint8(2), g.variantState.checksGiven[whiteStateIndex])
	require.Equal(t, uint8(1), clone.variantState.checksGiven[whiteStateIndex])
}
