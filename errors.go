package chessongo

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidFEN identifies invalid FEN input.
	ErrInvalidFEN = errors.New("invalid fen")
	// ErrIllegalMove identifies legal move rejection.
	ErrIllegalMove = errors.New("illegal move")
	// ErrInvalidMoveNotation identifies malformed move notation.
	ErrInvalidMoveNotation = errors.New("invalid move notation")
)

// FENField identifies the FEN field that failed validation.
type FENField string

const (
	FENFieldFormat         FENField = "format"
	FENFieldPiecePlacement FENField = "piece placement"
	FENFieldSideToMove     FENField = "side to move"
	FENFieldCastling       FENField = "castling"
	FENFieldEnPassant      FENField = "en passant"
	FENFieldHalfMoveClock  FENField = "halfmove clock"
	FENFieldFullMoveNumber FENField = "fullmove number"
	FENFieldVariantState   FENField = "variant state"
	FENFieldLegality       FENField = "legality"
)

// FENError provides structured context for FEN validation failures.
type FENError struct {
	Field  FENField
	Reason string
}

func (e *FENError) Error() string {
	if e == nil || e.Reason == "" {
		return ErrInvalidFEN.Error()
	}
	if e.Field != "" {
		return fmt.Sprintf("%s: %s: %s", ErrInvalidFEN, e.Field, e.Reason)
	}
	return fmt.Sprintf("%s: %s", ErrInvalidFEN, e.Reason)
}

func (e *FENError) Unwrap() error {
	return ErrInvalidFEN
}

func invalidFEN(reason string) error {
	return &FENError{Field: FENFieldFormat, Reason: reason}
}

func invalidFENField(field FENField, reason string) error {
	return &FENError{Field: field, Reason: reason}
}

// IllegalMoveReason identifies why a move request was rejected.
type IllegalMoveReason string

const (
	IllegalMoveReasonNotLegal            IllegalMoveReason = "not legal"
	IllegalMoveReasonNoMatchingLegalMove IllegalMoveReason = "no matching legal move"
)

// IllegalMoveError provides structured context for legal move rejection.
type IllegalMoveError struct {
	Move     Move
	Notation string
	Reason   IllegalMoveReason
}

func (e *IllegalMoveError) Error() string {
	if e == nil {
		return ErrIllegalMove.Error()
	}
	reason := string(e.Reason)
	if reason == "" {
		reason = string(IllegalMoveReasonNotLegal)
	}
	switch {
	case e.Notation != "":
		return fmt.Sprintf("%s: %s: %s", ErrIllegalMove, e.Notation, reason)
	case e.Move != 0:
		return fmt.Sprintf("%s: %s: %s", ErrIllegalMove, e.Move.UCI(), reason)
	default:
		return fmt.Sprintf("%s: %s", ErrIllegalMove, reason)
	}
}

func (e *IllegalMoveError) Unwrap() error {
	return ErrIllegalMove
}

func illegalMove(move Move, notation string, reason IllegalMoveReason) error {
	return &IllegalMoveError{Move: move, Notation: notation, Reason: reason}
}
