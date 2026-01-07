package server

import (
	"bataille-navale/game"
	"bataille-navale/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestServer() *Server {
	g := game.NewGame(10, []int{5, 4, 3})
	return NewServer(g)
}

func TestHandleBoard(t *testing.T) {
	s := setupTestServer()
	
	req := httptest.NewRequest(http.MethodGet, "/board", nil)
	w := httptest.NewRecorder()
	
	s.HandleBoard(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var resp models.BoardResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if resp.Size != 10 {
		t.Errorf("Expected board size 10, got %d", resp.Size)
	}
	
	if len(resp.Cells) != 10 {
		t.Errorf("Expected 10 rows, got %d", len(resp.Cells))
	}
}

func TestHandleBoard_MethodNotAllowed(t *testing.T) {
	s := setupTestServer()
	
	req := httptest.NewRequest(http.MethodPost, "/board", nil)
	w := httptest.NewRecorder()
	
	s.HandleBoard(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleBoats(t *testing.T) {
	s := setupTestServer()
	
	req := httptest.NewRequest(http.MethodGet, "/boats", nil)
	w := httptest.NewRecorder()
	
	s.HandleBoats(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var resp models.BoatsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if resp.RemainingBoats != 3 {
		t.Errorf("Expected 3 boats, got %d", resp.RemainingBoats)
	}
}

func TestHandleHit_ValidCoordinates(t *testing.T) {
	s := setupTestServer()
	
	hitReq := models.HitRequest{X: 5, Y: 5}
	body, _ := json.Marshal(hitReq)
	
	req := httptest.NewRequest(http.MethodPost, "/hit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	s.HandleHit(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var resp models.HitResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if resp.X != 5 || resp.Y != 5 {
		t.Errorf("Expected coords (5,5), got (%d,%d)", resp.X, resp.Y)
	}
	
	if resp.Result != "hit" && resp.Result != "miss" {
		t.Errorf("Expected 'hit' or 'miss', got '%s'", resp.Result)
	}
}

func TestHandleHit_InvalidCoordinates(t *testing.T) {
	s := setupTestServer()
	
	tests := []struct {
		name string
		x    int
		y    int
	}{
		{"X too high", 15, 5},
		{"Y too high", 5, 15},
		{"X negative", -1, 5},
		{"Y negative", 5, -1},
		{"Both invalid", 100, -50},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hitReq := models.HitRequest{X: tt.x, Y: tt.y}
			body, _ := json.Marshal(hitReq)
			
			req := httptest.NewRequest(http.MethodPost, "/hit", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			s.HandleHit(w, req)
			
			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", w.Code)
			}
			
			var errResp map[string]string
			json.NewDecoder(w.Body).Decode(&errResp)
			
			if _, ok := errResp["error"]; !ok {
				t.Error("Expected error field in response")
			}
		})
	}
}

func TestHandleHit_InvalidJSON(t *testing.T) {
	s := setupTestServer()
	
	req := httptest.NewRequest(http.MethodPost, "/hit", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	s.HandleHit(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleHit_MissingContentType(t *testing.T) {
	s := setupTestServer()
	
	hitReq := models.HitRequest{X: 5, Y: 5}
	body, _ := json.Marshal(hitReq)
	
	req := httptest.NewRequest(http.MethodPost, "/hit", bytes.NewBuffer(body))
	// Pas de Content-Type
	w := httptest.NewRecorder()
	
	s.HandleHit(w, req)
	
	// Devrait accepter (ContentType vide est permis dans validation)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleHits(t *testing.T) {
	s := setupTestServer()
	
	// D'abord envoyer un hit
	hitReq := models.HitRequest{X: 3, Y: 3}
	body, _ := json.Marshal(hitReq)
	reqHit := httptest.NewRequest(http.MethodPost, "/hit", bytes.NewBuffer(body))
	reqHit.Header.Set("Content-Type", "application/json")
	wHit := httptest.NewRecorder()
	s.HandleHit(wHit, reqHit)
	
	// Puis récupérer les hits
	req := httptest.NewRequest(http.MethodGet, "/hits", nil)
	w := httptest.NewRecorder()
	
	s.HandleHits(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var resp models.HitsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if len(resp.Hits) == 0 {
		t.Error("Expected at least 1 hit in history")
	}
}

func TestHandleSpecialShot_Cross(t *testing.T) {
	s := setupTestServer()
	
	shotReq := models.ShotRequest{
		X:        5,
		Y:        5,
		ShotType: models.ShotCross,
	}
	body, _ := json.Marshal(shotReq)
	
	req := httptest.NewRequest(http.MethodPost, "/special-shot", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	s.HandleSpecialShot(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var resp models.ShotResult
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if resp.ShotType != models.ShotCross {
		t.Errorf("Expected ShotCross, got %s", resp.ShotType)
	}
	
	if len(resp.Hits) == 0 {
		t.Error("Expected at least 1 hit")
	}
}

func TestHandleSpecialShot_InvalidType(t *testing.T) {
	s := setupTestServer()
	
	shotReq := models.ShotRequest{
		X:        5,
		Y:        5,
		ShotType: "invalid_type",
	}
	body, _ := json.Marshal(shotReq)
	
	req := httptest.NewRequest(http.MethodPost, "/special-shot", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	s.HandleSpecialShot(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleRegister(t *testing.T) {
	s := setupTestServer()
	
	registered := false
	s.OnOpponentAdded = func(name, url string) {
		if name == "TestPlayer" && url == "http://localhost:8080" {
			registered = true
		}
	}
	
	regReq := models.RegisterRequest{
		Name: "TestPlayer",
		URL:  "http://localhost:8080",
	}
	body, _ := json.Marshal(regReq)
	
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	s.HandleRegister(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if !registered {
		t.Error("OnOpponentAdded callback was not called")
	}
}

func TestHandleRegister_MissingFields(t *testing.T) {
	s := setupTestServer()
	
	tests := []struct {
		name string
		req  models.RegisterRequest
	}{
		{"Missing name", models.RegisterRequest{URL: "http://localhost:8080"}},
		{"Missing URL", models.RegisterRequest{Name: "Player"}},
		{"Both missing", models.RegisterRequest{}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.req)
			
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			s.HandleRegister(w, req)
			
			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", w.Code)
			}
		})
	}
}

func TestHandle404(t *testing.T) {
	s := setupTestServer()
	mux := http.NewServeMux()
	s.SetupRoutes(mux)
	
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	
	mux.ServeHTTP(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
	
	var errResp map[string]string
	json.NewDecoder(w.Body).Decode(&errResp)
	
	if _, ok := errResp["error"]; !ok {
		t.Error("Expected error field in 404 response")
	}
}

func TestRootRoute(t *testing.T) {
	s := setupTestServer()
	mux := http.NewServeMux()
	s.SetupRoutes(mux)
	
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	
	mux.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	
	if resp["message"] != "Bataille Navale Server" {
		t.Errorf("Unexpected message: %s", resp["message"])
	}
}

func TestConcurrentRequests(t *testing.T) {
	s := setupTestServer()
	
	done := make(chan bool)
	
	// Simuler 50 requêtes concurrentes
	for i := 0; i < 50; i++ {
		go func(x, y int) {
			hitReq := models.HitRequest{X: x % 10, Y: y % 10}
			body, _ := json.Marshal(hitReq)
			
			req := httptest.NewRequest(http.MethodPost, "/hit", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			s.HandleHit(w, req)
			
			done <- true
		}(i, i)
	}
	
	// Attendre toutes les requêtes
	for i := 0; i < 50; i++ {
		<-done
	}
	
	// Vérifier que le serveur est toujours cohérent
	req := httptest.NewRequest(http.MethodGet, "/boats", nil)
	w := httptest.NewRecorder()
	s.HandleBoats(w, req)
	
	if w.Code != http.StatusOK {
		t.Error("Server state corrupted after concurrent requests")
	}
}