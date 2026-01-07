package main

import (
	"bataille-navale/client"
	"bataille-navale/game"
	"bataille-navale/models"
	"bataille-navale/server"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupTestServerWithURL() (*server.Server, string) {
	g := game.NewGame(10, []int{5, 4, 3})
	g.DisableMovement = true // Désactiver mouvement pour tests
	s := server.NewServer(g)
	
	mux := http.NewServeMux()
	s.SetupRoutes(mux)
	
	ts := httptest.NewServer(mux)
	
	return s, ts.URL
}

func TestClientServerIntegration_GetBoard(t *testing.T) {
	_, url := setupTestServerWithURL()
	
	c := client.NewClient(url)
	
	board, err := c.GetBoard()
	if err != nil {
		t.Fatalf("Failed to get board: %v", err)
	}
	
	if board.Size != 10 {
		t.Errorf("Expected board size 10, got %d", board.Size)
	}
	
	if len(board.Cells) != 10 {
		t.Errorf("Expected 10 rows, got %d", len(board.Cells))
	}
}

func TestClientServerIntegration_GetBoatsCount(t *testing.T) {
	_, url := setupTestServerWithURL()
	
	c := client.NewClient(url)
	
	boats, err := c.GetBoatsCount()
	if err != nil {
		t.Fatalf("Failed to get boats count: %v", err)
	}
	
	if boats != 3 {
		t.Errorf("Expected 3 boats, got %d", boats)
	}
}

func TestClientServerIntegration_Hit(t *testing.T) {
	_, url := setupTestServerWithURL()
	
	c := client.NewClient(url)
	
	resp, err := c.Hit(5, 5)
	if err != nil {
		t.Fatalf("Failed to hit: %v", err)
	}
	
	if resp.X != 5 || resp.Y != 5 {
		t.Errorf("Expected coords (5,5), got (%d,%d)", resp.X, resp.Y)
	}
	
	if resp.Result != "hit" && resp.Result != "miss" {
		t.Errorf("Expected 'hit' or 'miss', got '%s'", resp.Result)
	}
}

func TestClientServerIntegration_InvalidCoordinates(t *testing.T) {
	_, url := setupTestServerWithURL()
	
	c := client.NewClient(url)
	
	_, err := c.Hit(15, 15)
	if err == nil {
		t.Error("Expected error for invalid coordinates")
	}
}

func TestClientServerIntegration_GetHits(t *testing.T) {
	_, url := setupTestServerWithURL()
	
	c := client.NewClient(url)
	
	// Faire quelques tirs
	c.Hit(1, 1)
	c.Hit(2, 2)
	c.Hit(3, 3)
	
	hits, err := c.GetHits()
	if err != nil {
		t.Fatalf("Failed to get hits: %v", err)
	}
	
	if len(hits.Hits) < 3 {
		t.Errorf("Expected at least 3 hits, got %d", len(hits.Hits))
	}
}

func TestClientServerIntegration_IsAlive(t *testing.T) {
	_, url := setupTestServerWithURL()
	
	c := client.NewClient(url)
	
	if !c.IsAlive() {
		t.Error("Server should be alive initially")
	}
}

func TestClientServerIntegration_SpecialShot(t *testing.T) {
	_, url := setupTestServerWithURL()
	
	c := client.NewClient(url)
	
	result, err := c.SpecialShot(5, 5, models.ShotCross)
	if err != nil {
		t.Fatalf("Failed to special shot: %v", err)
	}
	
	if result.ShotType != models.ShotCross {
		t.Errorf("Expected ShotCross, got %s", result.ShotType)
	}
	
	if len(result.Hits) == 0 {
		t.Error("Expected at least 1 hit")
	}
}

func TestClientServerIntegration_Register(t *testing.T) {
	s, url := setupTestServerWithURL()
	
	registered := false
	s.OnOpponentAdded = func(name, url string) {
		if name == "TestPlayer" {
			registered = true
		}
	}
	
	c := client.NewClient(url)
	
	err := c.Register("TestPlayer", "http://localhost:9999")
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}
	
	time.Sleep(100 * time.Millisecond) // Attendre le callback
	
	if !registered {
		t.Error("Registration callback was not triggered")
	}
}

func TestClientServerIntegration_FullGame(t *testing.T) {
	_, url := setupTestServerWithURL()
	
	c := client.NewClient(url)
	
	// Phase 1: Vérifier l'état initial
	boats, _ := c.GetBoatsCount()
	if boats != 3 {
		t.Errorf("Expected 3 boats initially, got %d", boats)
	}
	
	// Phase 2: Effectuer plusieurs tirs
	hitsCount := 0
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			resp, err := c.Hit(x, y)
			if err != nil {
				t.Fatalf("Hit failed at (%d,%d): %v", x, y, err)
			}
			
			if resp.Result == "hit" {
				hitsCount++
			}
		}
	}
	
	t.Logf("Total hits: %d", hitsCount)
	
	// Phase 3: Vérifier que tous les bateaux sont coulés
	boats, _ = c.GetBoatsCount()
	if boats != 0 {
		t.Errorf("Expected 0 boats after full scan, got %d", boats)
	}
	
	// Phase 4: Vérifier IsAlive
	if c.IsAlive() {
		t.Error("Server should not be alive after all boats sunk")
	}
}

func TestClientServerIntegration_Retry(t *testing.T) {
	// Créer un serveur qui échoue 2 fois puis réussit
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// 3ème tentative réussit
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"remaining_boats": 5}`))
	}))
	defer ts.Close()
	
	c := client.NewClient(ts.URL)
	c.MaxRetries = 3
	c.RetryDelay = 100 * time.Millisecond
	
	boats, err := c.GetBoatsCount()
	if err != nil {
		t.Fatalf("Should succeed after retries: %v", err)
	}
	
	if boats != 5 {
		t.Errorf("Expected 5 boats, got %d", boats)
	}
	
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestClientServerIntegration_RetryFailure(t *testing.T) {
	// Serveur qui échoue toujours
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()
	
	c := client.NewClient(ts.URL)
	c.MaxRetries = 2
	c.RetryDelay = 50 * time.Millisecond
	
	_, err := c.GetBoatsCount()
	if err == nil {
		t.Error("Expected error after max retries")
	}
}

func TestClientServerIntegration_Concurrent(t *testing.T) {
	_, url := setupTestServerWithURL()
	
	c := client.NewClient(url)
	
	done := make(chan bool)
	errors := 0
	
	// 20 clients concurrents
	for i := 0; i < 20; i++ {
		go func(x, y int) {
			_, err := c.Hit(x%10, y%10)
			if err != nil {
				errors++
			}
			done <- true
		}(i, i)
	}
	
	// Attendre tous les clients
	for i := 0; i < 20; i++ {
		<-done
	}
	
	if errors > 0 {
		t.Errorf("Got %d errors during concurrent access", errors)
	}
	
	// Vérifier que le serveur est toujours cohérent
	boats, err := c.GetBoatsCount()
	if err != nil {
		t.Error("Server corrupted after concurrent access")
	}
	
	t.Logf("Remaining boats after concurrent hits: %d", boats)
}

func TestClientServerIntegration_MultipleClients(t *testing.T) {
	_, url := setupTestServerWithURL()
	
	// Créer 5 clients
	clients := make([]*client.Client, 5)
	for i := 0; i < 5; i++ {
		clients[i] = client.NewClient(url)
	}
	
	// Chaque client tire 10 fois
	done := make(chan bool)
	
	for i, c := range clients {
		go func(clientID int, cl *client.Client) {
			for j := 0; j < 10; j++ {
				cl.Hit(j, clientID)
			}
			done <- true
		}(i, c)
	}
	
	// Attendre tous les clients
	for i := 0; i < 5; i++ {
		<-done
	}
	
	// Vérifier l'état final
	hits, err := clients[0].GetHits()
	if err != nil {
		t.Fatalf("Failed to get hits: %v", err)
	}
	
	if len(hits.Hits) < 50 {
		t.Errorf("Expected at least 50 hits from 5 clients, got %d", len(hits.Hits))
	}
	
	t.Logf("Total hits received: %d", len(hits.Hits))
}