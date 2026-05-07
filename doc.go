// Package chessongo provides standard chess rules, legal move generation, FEN
// loading, PGN import/export, undo support, and search-oriented traversal
// helpers.
//
// New code should prefer the stable method API such as Game.FEN,
// Game.SideToMove, Game.BoardView, Game.Snapshot, Game.LegalMovesInto,
// Game.DrawStatus, Game.PositionKey, Game.TryMoveUCI, Game.TryMoveSAN,
// Game.Clone, and Game.Status instead of mutating Game fields directly. Use
// ParseSquare and PieceFromRune instead of package-level lookup maps for new
// code.
package chessongo
