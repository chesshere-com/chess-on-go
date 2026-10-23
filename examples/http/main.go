package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	chessongo "github.com/chesshere-com/chess-on-go"
)

type moveRequest struct {
	FEN string `json:"fen"`
	UCI string `json:"uci"`
}

type moveResponse struct {
	FEN        string   `json:"fen"`
	Turn       string   `json:"turn"`
	LegalMoves []string `json:"legalMoves"`
}

func moveHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	defer r.Body.Close()
	var request moveRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	// One game per request. Do not mutate a shared Game concurrently.
	game, err := chessongo.NewGameFromFEN(request.FEN)
	if err != nil {
		http.Error(w, "Invalid FEN", http.StatusBadRequest)
		return
	}
	if err := game.TryMoveUCI(request.UCI); err != nil {
		http.Error(w, "Illegal move", http.StatusUnprocessableEntity)
		return
	}
	response := moveResponse{FEN: game.FEN(), Turn: game.SideToMove().String(), LegalMoves: []string{}}
	for _, move := range game.LegalMovesInto(nil) {
		response.LegalMoves = append(response.LegalMoves, move.UCI())
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Print(err)
	}
}

func main() {
	http.HandleFunc("POST /move", moveHandler)
	fmt.Println("Listening on http://127.0.0.1:8080")
	server := &http.Server{
		Addr:              "127.0.0.1:8080",
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
