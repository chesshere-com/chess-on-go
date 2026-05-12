package chessongo

import "testing"

func TestSeedPositionHistory_EmptyIsNoop(t *testing.T) {
	g := NewGame()
	g.SeedPositionHistory(nil)
	if g.CanClaimThreefoldRepetition() {
		t.Errorf("CanClaim should be false after nil seed on fresh game")
	}
	g.SeedPositionHistory([]uint64{})
	if g.CanClaimThreefoldRepetition() {
		t.Errorf("CanClaim should be false after empty seed on fresh game")
	}
}

func TestSeedPositionHistory_TwoPriorOccurrencesEnablesThreefoldClaim(t *testing.T) {
	g := NewGame()
	key := g.PositionKey()
	// FEN load recorded the current position once.
	// Seed two more occurrences → total count 3 → claimable.
	g.SeedPositionHistory([]uint64{key, key})
	if !g.CanClaimThreefoldRepetition() {
		t.Errorf("expected CanClaimThreefoldRepetition=true after seeding 2 prior copies of current key")
	}
	if g.IsFivefoldRepetition() {
		t.Errorf("expected IsFivefoldRepetition=false at count 3")
	}
}

func TestSeedPositionHistory_FourPriorOccurrencesTriggersFivefoldTerminal(t *testing.T) {
	g := NewGame()
	key := g.PositionKey()
	// Seed four more occurrences → total count 5 → automatic draw.
	g.SeedPositionHistory([]uint64{key, key, key, key})
	if !g.IsFivefoldRepetition() {
		t.Errorf("expected IsFivefoldRepetition=true after seeding 4 prior copies of current key")
	}
	if !g.IsTerminal() {
		t.Errorf("expected IsTerminal=true once fivefold reached (refreshStatus must run)")
	}
	if got := g.Status(); got != GameStatusDrawFivefoldRepetition {
		t.Errorf("Status() = %v, want GameStatusDrawFivefoldRepetition", got)
	}
}

func TestSeedPositionHistory_UnrelatedKeysDoNotTriggerClaim(t *testing.T) {
	g := NewGame()
	g.SeedPositionHistory([]uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	if g.CanClaimThreefoldRepetition() {
		t.Errorf("seeding keys unrelated to the current position must not trigger claim")
	}
	if g.IsFivefoldRepetition() {
		t.Errorf("seeding keys unrelated to the current position must not trigger fivefold")
	}
}

// TestSeedPositionHistory_LivePlayEquivalence is the key correctness guarantee:
// a Game reconstructed from a final FEN plus the chronological list of prior
// position keys must end up with an identical positionHistory map (and
// identical rule-detection answers) to a Game that played the same moves
// live. This is the exact use case the API exists to support.
func TestSeedPositionHistory_LivePlayEquivalence(t *testing.T) {
	// A knight shuffle that produces a threefold position on the final move.
	// White: g1-f3, f3-g1, g1-f3, f3-g1, g1-f3
	// Black: g8-f6, f6-g8, g8-f6, f6-g8
	ucis := []string{
		"g1f3", "g8f6",
		"f3g1", "f6g8",
		"g1f3", "g8f6",
		"f3g1", "f6g8",
		"g1f3",
	}

	live := NewGame()
	priorKeys := make([]uint64, 0, len(ucis))
	for _, uci := range ucis {
		// Record the BEFORE-move key — that's the key the library increments
		// on its next recordPosition call after replaying that move.
		priorKeys = append(priorKeys, live.PositionKey())
		if err := live.TryMoveUCI(uci); err != nil {
			t.Fatalf("live TryMoveUCI(%q): %v", uci, err)
		}
	}

	// Rebuild from final FEN + the prior keys.
	seeded, err := NewGameFromFEN(live.FEN())
	if err != nil {
		t.Fatalf("NewGameFromFEN: %v", err)
	}
	seeded.SeedPositionHistory(priorKeys)

	// positionHistory maps must agree exactly.
	if len(live.positionHistory) != len(seeded.positionHistory) {
		t.Fatalf("positionHistory size mismatch: live=%d seeded=%d",
			len(live.positionHistory), len(seeded.positionHistory))
	}
	for k, want := range live.positionHistory {
		if got := seeded.positionHistory[k]; got != want {
			t.Errorf("positionHistory[%x] mismatch: live=%d seeded=%d", k, want, got)
		}
	}

	// Rule-detection must agree.
	if got, want := seeded.CanClaimThreefoldRepetition(), live.CanClaimThreefoldRepetition(); got != want {
		t.Errorf("CanClaimThreefoldRepetition: live=%v seeded=%v", want, got)
	}
	if got, want := seeded.IsFivefoldRepetition(), live.IsFivefoldRepetition(); got != want {
		t.Errorf("IsFivefoldRepetition: live=%v seeded=%v", want, got)
	}
	if got, want := seeded.IsTerminal(), live.IsTerminal(); got != want {
		t.Errorf("IsTerminal: live=%v seeded=%v", want, got)
	}
	if got, want := seeded.Status(), live.Status(); got != want {
		t.Errorf("Status: live=%v seeded=%v", want, got)
	}

	// Sanity: this scenario reaches threefold on the live game.
	if !live.CanClaimThreefoldRepetition() {
		t.Fatalf("test scenario is broken — live game never reached threefold")
	}
}
