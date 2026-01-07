package main

import (
	"bataille-navale/client"
	"bataille-navale/game"
	"bataille-navale/models"
	"bataille-navale/server"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// setupLoadTestServer crée un serveur pour les tests de charge
func setupLoadTestServer() (*server.Server, string) {
	g := game.NewGame(10, []int{5, 4, 3, 3, 2})
	s := server.NewServer(g)
	
	mux := http.NewServeMux()
	s.SetupRoutes(mux)
	
	ts := httptest.NewServer(mux)
	
	return s, ts.URL
}

// Test de charge: 100 requêtes séquentielles
func TestLoad_100SequentialRequests(t *testing.T) {
	_, url := setupLoadTestServer()
	c := client.NewClient(url)
	
	start := time.Now()
	
	for i := 0; i < 100; i++ {
		_, err := c.GetBoatsCount()
		if err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}
	}
	
	duration := time.Since(start)
	t.Logf("100 sequential requests completed in %v (avg: %v per request)", 
		duration, duration/100)
}

// Test de charge: 100 requêtes concurrentes
func TestLoad_100ConcurrentRequests(t *testing.T) {
	_, url := setupLoadTestServer()
	
	var wg sync.WaitGroup
	var successCount, errorCount int32
	
	start := time.Now()
	
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			c := client.NewClient(url)
			_, err := c.GetBoatsCount()
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
			} else {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}
	
	wg.Wait()
	duration := time.Since(start)
	
	t.Logf("100 concurrent requests completed in %v", duration)
	t.Logf("Success: %d, Errors: %d", successCount, errorCount)
	
	if errorCount > 0 {
		t.Errorf("Got %d errors out of 100 requests", errorCount)
	}
}

// Test de charge: 500 tirs concurrents
func TestLoad_500ConcurrentHits(t *testing.T) {
	_, url := setupLoadTestServer()
	
	var wg sync.WaitGroup
	var hitCount, missCount, errorCount int32
	
	start := time.Now()
	
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			c := client.NewClient(url)
			resp, err := c.Hit(id%10, (id/10)%10)
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
			} else if resp.Result == "hit" {
				atomic.AddInt32(&hitCount, 1)
			} else {
				atomic.AddInt32(&missCount, 1)
			}
		}(i)
	}
	
	wg.Wait()
	duration := time.Since(start)
	
	t.Logf("500 concurrent hits completed in %v", duration)
	t.Logf("Hits: %d, Misses: %d, Errors: %d", hitCount, missCount, errorCount)
	
	if errorCount > 0 {
		t.Errorf("Got %d errors out of 500 requests", errorCount)
	}
}

// Test de charge: Plusieurs joueurs attaquant simultanément
func TestLoad_MultiplayerSimulation(t *testing.T) {
	numPlayers := 10
	hitsPerPlayer := 50
	
	// Créer un serveur cible
	_, targetURL := setupLoadTestServer()
	
	var wg sync.WaitGroup
	var totalHits, totalMisses, totalErrors int32
	
	start := time.Now()
	
	// Chaque joueur tire 50 fois
	for i := 0; i < numPlayers; i++ {
		wg.Add(1)
		go func(playerID int) {
			defer wg.Done()
			
			c := client.NewClient(targetURL)
			
			for j := 0; j < hitsPerPlayer; j++ {
				x := (playerID + j) % 10
				y := (playerID*j) % 10
				
				resp, err := c.Hit(x, y)
				if err != nil {
					atomic.AddInt32(&totalErrors, 1)
				} else if resp.Result == "hit" {
					atomic.AddInt32(&totalHits, 1)
				} else {
					atomic.AddInt32(&totalMisses, 1)
				}
			}
		}(i)
	}
	
	wg.Wait()
	duration := time.Since(start)
	
	totalRequests := numPlayers * hitsPerPlayer
	t.Logf("%d players x %d hits = %d total requests", numPlayers, hitsPerPlayer, totalRequests)
	t.Logf("Completed in %v (avg: %v per request)", 
		duration, duration/time.Duration(totalRequests))
	t.Logf("Results - Hits: %d, Misses: %d, Errors: %d", 
		totalHits, totalMisses, totalErrors)
	
	if totalErrors > 0 {
		t.Errorf("Got %d errors out of %d requests", totalErrors, totalRequests)
	}
}

// Test de charge: Tirs spéciaux concurrents
func TestLoad_SpecialShotsConcurrent(t *testing.T) {
	_, url := setupLoadTestServer()
	
	shotTypes := []models.ShotType{
		models.ShotNormal,
		models.ShotCross,
		models.ShotZone,
		models.ShotParalyzing,
	}
	
	var wg sync.WaitGroup
	var successCount, errorCount int32
	
	start := time.Now()
	
	// 200 tirs spéciaux aléatoires
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			c := client.NewClient(url)
			shotType := shotTypes[id%len(shotTypes)]
			
			_, err := c.SpecialShot(id%10, (id/10)%10, shotType)
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
			} else {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}
	
	wg.Wait()
	duration := time.Since(start)
	
	t.Logf("200 concurrent special shots completed in %v", duration)
	t.Logf("Success: %d, Errors: %d", successCount, errorCount)
	
	if errorCount > 0 {
		t.Errorf("Got %d errors out of 200 requests", errorCount)
	}
}

// Test de charge: Stress test avec burst de requêtes
func TestLoad_BurstRequests(t *testing.T) {
	_, url := setupLoadTestServer()
	
	bursts := 5
	requestsPerBurst := 100
	
	start := time.Now()
	
	for burst := 0; burst < bursts; burst++ {
		var wg sync.WaitGroup
		var errorCount int32
		
		burstStart := time.Now()
		
		for i := 0; i < requestsPerBurst; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				c := client.NewClient(url)
				
				// Mélange de requêtes
				switch id % 4 {
				case 0:
					_, err := c.GetBoatsCount()
					if err != nil {
						atomic.AddInt32(&errorCount, 1)
					}
				case 1:
					_, err := c.GetBoard()
					if err != nil {
						atomic.AddInt32(&errorCount, 1)
					}
				case 2:
					_, err := c.Hit(id%10, (id/10)%10)
					if err != nil {
						atomic.AddInt32(&errorCount, 1)
					}
				case 3:
					_, err := c.GetHits()
					if err != nil {
						atomic.AddInt32(&errorCount, 1)
					}
				}
			}(i)
		}
		
		wg.Wait()
		burstDuration := time.Since(burstStart)
		
		t.Logf("Burst %d: %d requests in %v (errors: %d)", 
			burst+1, requestsPerBurst, burstDuration, errorCount)
		
		if errorCount > 0 {
			t.Errorf("Burst %d had %d errors", burst+1, errorCount)
		}
		
		// Pause entre les bursts
		time.Sleep(100 * time.Millisecond)
	}
	
	totalDuration := time.Since(start)
	t.Logf("Total burst test duration: %v", totalDuration)
}

// Test de charge: Simulation de jeu complet
func TestLoad_FullGameSimulation(t *testing.T) {
	_, url := setupLoadTestServer()
	
	numPlayers := 5
	var wg sync.WaitGroup
	
	start := time.Now()
	
	for playerID := 0; playerID < numPlayers; playerID++ {
		wg.Add(1)
		go func(pid int) {
			defer wg.Done()
			
			c := client.NewClient(url)
			
			// Phase 1: Récupérer l'état initial
			_, err := c.GetBoatsCount()
			if err != nil {
				t.Errorf("Player %d failed to get boats: %v", pid, err)
				return
			}
			
			// Phase 2: Scanner le plateau
			for y := 0; y < 10; y++ {
				for x := 0; x < 10; x++ {
					// Seulement certaines cases pour ne pas trop charger
					if (x+y+pid)%3 == 0 {
						_, err := c.Hit(x, y)
						if err != nil {
							t.Errorf("Player %d hit failed at (%d,%d): %v", pid, x, y, err)
						}
					}
				}
			}
			
			// Phase 3: Quelques tirs spéciaux
			c.SpecialShot(5, 5, models.ShotCross)
			c.SpecialShot(3, 3, models.ShotZone)
			
			// Phase 4: Vérifier l'état final
			_, err = c.GetHits()
			if err != nil {
				t.Errorf("Player %d failed to get hits: %v", pid, err)
			}
			
			t.Logf("Player %d completed simulation", pid)
		}(playerID)
	}
	
	wg.Wait()
	duration := time.Since(start)
	
	t.Logf("Full game simulation with %d players completed in %v", numPlayers, duration)
}

// Test de charge: Mesure de throughput
func TestLoad_ThroughputMeasurement(t *testing.T) {
	_, url := setupLoadTestServer()
	
	duration := 5 * time.Second
	var requestCount int64
	
	stop := make(chan bool)
	var wg sync.WaitGroup
	
	// Lancer 10 workers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := client.NewClient(url)
			
			for {
				select {
				case <-stop:
					return
				default:
					c.GetBoatsCount()
					atomic.AddInt64(&requestCount, 1)
				}
			}
		}()
	}
	
	// Laisser tourner pendant X secondes
	time.Sleep(duration)
	close(stop)
	wg.Wait()
	
	throughput := float64(requestCount) / duration.Seconds()
	t.Logf("Throughput: %.2f requests/second (%d total requests in %v)", 
		throughput, requestCount, duration)
}

// Test de charge: Validation de la cohérence sous charge
func TestLoad_ConsistencyUnderLoad(t *testing.T) {
	_, url := setupLoadTestServer()
	
	c := client.NewClient(url)
	
	// État initial
	initialBoats, _ := c.GetBoatsCount()
	
	// 1000 tirs aléatoires concurrents
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			client := client.NewClient(url)
			client.Hit(id%10, (id/10)%10)
		}(i)
	}
	wg.Wait()
	
	// Vérifier l'état final
	finalBoats, err := c.GetBoatsCount()
	if err != nil {
		t.Fatalf("Failed to get final boat count: %v", err)
	}
	
	// Le nombre de bateaux ne devrait jamais augmenter
	if finalBoats > initialBoats {
		t.Errorf("Inconsistent state: boats increased from %d to %d", initialBoats, finalBoats)
	}
	
	// Le nombre de bateaux devrait être entre 0 et initialBoats
	if finalBoats < 0 || finalBoats > initialBoats {
		t.Errorf("Invalid boat count: %d (should be between 0 and %d)", finalBoats, initialBoats)
	}
	
	t.Logf("Consistency check passed: %d -> %d boats", initialBoats, finalBoats)
}

// Benchmark: Performance d'un hit simple
func BenchmarkHit(b *testing.B) {
	_, url := setupLoadTestServer()
	c := client.NewClient(url)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Hit(i%10, (i/10)%10)
	}
}

// Benchmark: Performance d'un GetBoatsCount
func BenchmarkGetBoatsCount(b *testing.B) {
	_, url := setupLoadTestServer()
	c := client.NewClient(url)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.GetBoatsCount()
	}
}

// Benchmark: Performance d'un tir spécial (Cross)
func BenchmarkSpecialShotCross(b *testing.B) {
	_, url := setupLoadTestServer()
	c := client.NewClient(url)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SpecialShot(5, 5, models.ShotCross)
	}
}

// Benchmark: Performance d'un tir spécial (Zone)
func BenchmarkSpecialShotZone(b *testing.B) {
	_, url := setupLoadTestServer()
	c := client.NewClient(url)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SpecialShot(5, 5, models.ShotZone)
	}
}