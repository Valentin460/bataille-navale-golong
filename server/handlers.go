package server

import (
	"bataille-navale/game"
	"bataille-navale/models"
	"net/http"
)

type Server struct {
	Game *game.Game
	OnOpponentAdded func(name, url string) // Callback pour notifier le CLI
}

func NewServer(g *game.Game) *Server {
	return &Server{
		Game: g,
	}
}

func (s *Server) HandleBoard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	state := s.Game.GetBoardState()
	RespondJSON(w, http.StatusOK, state)
}

func (s *Server) HandleBoats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	response := models.BoatsResponse{
		RemainingBoats: s.Game.GetRemainingBoats(),
	}
	
	RespondJSON(w, http.StatusOK, response)
}

func (s *Server) HandleHit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	var req models.HitRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	// Validation des coordonnées
	if err := ValidateCoordinates(req.X, req.Y, s.Game.Board.Size); err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	response := s.Game.ProcessHit(req.X, req.Y)
	
	RespondJSON(w, http.StatusOK, response)
}

func (s *Server) HandleHits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	response := s.Game.GetReceivedHits()
	RespondJSON(w, http.StatusOK, response)
}

func (s *Server) HandleSpecialShot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	var req models.ShotRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	// Validation des coordonnées
	if err := ValidateCoordinates(req.X, req.Y, s.Game.Board.Size); err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	// Validation du type de tir
	if req.ShotType == "" {
		req.ShotType = models.ShotNormal
	}
	
	if !req.ShotType.IsValid() {
		RespondError(w, http.StatusBadRequest, "Invalid shot type")
		return
	}
	
	response := s.Game.ProcessSpecialShot(req.X, req.Y, req.ShotType)
	
	RespondJSON(w, http.StatusOK, response)
}

// HandleRegister permet à un adversaire de s'enregistrer mutuellement
func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	var req models.RegisterRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	// Validation des champs
	if req.Name == "" {
		RespondError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.URL == "" {
		RespondError(w, http.StatusBadRequest, "URL is required")
		return
	}
	
	// Notifier le CLI pour ajouter l'adversaire
	if s.OnOpponentAdded != nil {
		s.OnOpponentAdded(req.Name, req.URL)
	}
	
	RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Opponent registered successfully",
	})
}

// Handle404 gère les routes inexistantes
func (s *Server) Handle404(w http.ResponseWriter, r *http.Request) {
	RespondError(w, http.StatusNotFound, "Route not found: "+r.URL.Path)
}

func (s *Server) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/board", s.HandleBoard)
	mux.HandleFunc("/boats", s.HandleBoats)
	mux.HandleFunc("/hit", s.HandleHit)
	mux.HandleFunc("/hits", s.HandleHits)
	mux.HandleFunc("/special-shot", s.HandleSpecialShot)
	mux.HandleFunc("/register", s.HandleRegister)
	
	// Handler 404 pour toutes les autres routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			s.Handle404(w, r)
		} else {
			RespondJSON(w, http.StatusOK, map[string]string{
				"message": "Bataille Navale Server",
				"version": "1.0.0",
			})
		}
	})
}