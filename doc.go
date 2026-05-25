// Package chessongo provides chess rules, legal move generation, FEN loading,
// PGN import/export, undo support, and search-oriented traversal helpers.
//
// Standard chess is the default rule set. Chess960, King of the Hill, and
// Three-check are available through explicit variant constructors or
// LoadFENWithVariant.
//
// The stable method API includes Game.FEN, Game.SideToMove, Game.BoardView,
// Game.Snapshot, Game.LegalMovesInto, Game.DrawStatus, Game.PositionKey,
// Game.TryMoveUCI, Game.TryMoveSAN, Game.Clone, and Game.Status. Prefer
// ParseSquare and PieceFromRune over the package-level lookup maps.
package chessongo
