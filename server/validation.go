package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	MaxBodySize = 1 << 20 // 1 MB
)

// ValidateCoordinates vérifie que x,y sont dans les limites du plateau
func ValidateCoordinates(x, y, boardSize int) error {
	if x < 0 || x >= boardSize {
		return fmt.Errorf("coordonnée X invalide: %d (doit être entre 0 et %d)", x, boardSize-1)
	}
	if y < 0 || y >= boardSize {
		return fmt.Errorf("coordonnée Y invalide: %d (doit être entre 0 et %d)", y, boardSize-1)
	}
	return nil
}

// ValidateContentType vérifie que le Content-Type est application/json
func ValidateContentType(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" && contentType != "" {
		return fmt.Errorf("Content-Type invalide: attendu 'application/json', reçu '%s'", contentType)
	}
	return nil
}

// LimitBodySize limite la taille du body de la requête
func LimitBodySize(r *http.Request) {
	r.Body = http.MaxBytesReader(nil, r.Body, MaxBodySize)
}

// DecodeAndValidate décode le JSON et gère les erreurs
func DecodeAndValidate(r *http.Request, v interface{}) error {
	LimitBodySize(r)
	
	if err := ValidateContentType(r); err != nil {
		return err
	}
	
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Rejette les champs inconnus
	
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("erreur de décodage JSON: %w", err)
	}
	
	return nil
}

// RespondJSON envoie une réponse JSON
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// RespondError envoie une erreur JSON
func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, map[string]string{"error": message})
}