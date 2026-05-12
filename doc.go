// Package chessongo provides standard chess rules, legal move generation, FEN
// loading, PGN import/export, undo support, and search-oriented traversal
// helpers.
//
// The stable method API includes Game.FEN, Game.SideToMove, Game.BoardView,
// Game.Snapshot, Game.LegalMovesInto, Game.DrawStatus, Game.PositionKey,
// Game.TryMoveUCI, Game.TryMoveSAN, Game.Clone, and Game.Status. Prefer
// ParseSquare and PieceFromRune over the package-level lookup maps.
package chessongo
