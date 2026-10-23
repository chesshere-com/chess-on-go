package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chessongo "github.com/chesshere-com/chess-on-go"
)

func TestMoveHandler(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"legal", `{"fen":"` + chessongo.STARTING_POSITION_FEN + `","uci":"e2e4"}`, http.StatusOK},
		{"illegal", `{"fen":"` + chessongo.STARTING_POSITION_FEN + `","uci":"e2e5"}`, http.StatusUnprocessableEntity},
		{"invalid FEN", `{"fen":"bad","uci":"e2e4"}`, http.StatusBadRequest},
		{"invalid JSON", "{", http.StatusBadRequest},
		{"oversized", `{"fen":"` + strings.Repeat("x", 5000) + `","uci":"e2e4"}`, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/move", strings.NewReader(tc.body))
			moveHandler(response, request)
			if response.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, tc.status, response.Body.String())
			}
			if tc.status == http.StatusOK {
				var payload moveResponse
				if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
					t.Fatal(err)
				}
				if payload.Turn != "black" || len(payload.LegalMoves) != 20 || payload.FEN != "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1" {
					t.Fatalf("unexpected position: %+v", payload)
				}
			}
		})
	}
}
