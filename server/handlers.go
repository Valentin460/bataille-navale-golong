package server

import (
	"bataille-navale/game"
	"bataille-navale/models"
	"encoding/json"
	"net/http"
)

type Server struct {
	Game *game.Game
}

func NewServer(g *game.Game) *Server {
	return &Server{
		Game: g,
	}
}

func (s *Server) HandleBoard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	state := s.Game.GetBoardState()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (s *Server) HandleBoats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	response := models.BoatsResponse{
		RemainingBoats: s.Game.GetRemainingBoats(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) HandleHit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req models.HitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	response := s.Game.ProcessHit(req.X, req.Y)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) HandleHits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	response := s.Game.GetReceivedHits()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) HandleSpecialShot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req models.ShotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.ShotType == "" {
		req.ShotType = models.ShotNormal
	}
	
	if !req.ShotType.IsValid() {
		http.Error(w, "Invalid shot type", http.StatusBadRequest)
		return
	}
	
	response := s.Game.ProcessSpecialShot(req.X, req.Y, req.ShotType)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/board", s.HandleBoard)
	mux.HandleFunc("/boats", s.HandleBoats)
	mux.HandleFunc("/hit", s.HandleHit)
	mux.HandleFunc("/hits", s.HandleHits)
	mux.HandleFunc("/special-shot", s.HandleSpecialShot)
}
