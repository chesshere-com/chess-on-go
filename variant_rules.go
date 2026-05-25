package chessongo

type computedStatus struct {
	isCheckmate           bool
	isStalemate           bool
	isMaterialDraw        bool
	isThreefoldRepetition bool
	isFiftyMoveRule       bool
	isSeventyFiveMoveRule bool
	isFinished            bool
	winner                Color
	status                GameStatus
}

type variantRules struct {
	variant     Variant
	name        string
	implemented bool

	afterMove      func(g *Game, move Move, prev *GameState)
	overrideStatus func(g *Game, status *computedStatus)
	hashExtra      func(g *Game) uint64

	pgnVariantTag string
	pgnNames      []string
}

var unknownVariantRules = variantRules{
	variant: VariantStandard,
	name:    "unknown",
}

func rulesForVariant(variant Variant) variantRules {
	switch variant {
	case VariantStandard:
		return standardVariantRules
	case VariantChess960:
		return chess960VariantRules
	case VariantKingOfTheHill:
		return kingOfTheHillVariantRules
	case VariantThreeCheck:
		return threeCheckVariantRules
	default:
		return unknownVariantRules
	}
}

func (g *Game) rules() variantRules {
	return rulesForVariant(g.variant)
}
