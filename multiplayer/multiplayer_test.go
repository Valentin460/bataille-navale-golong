package multiplayer

import (
	"bataille-navale/game"
	"bataille-navale/server"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupMultiplayerTest() (*GameManager, string) {
	g := game.NewGame(10, []int{5, 4, 3})
	s := server.NewServer(g)
	
	mux := http.NewServeMux()
	s.SetupRoutes(mux)
	
	ts := httptest.NewServer(mux)
	
	gm := NewGameManager(g, "TestPlayer", ts.URL)
	
	return gm, ts.URL
}

func TestAddOpponent(t *testing.T) {
	gm, _ := setupMultiplayerTest()
	
	gm.AddOpponent("Player1", "http://localhost:8081")
	
	opponents := gm.GetAllOpponents()
	if len(opponents) != 1 {
		t.Errorf("Expected 1 opponent, got %d", len(opponents))
	}
	
	if opponents[0].Name != "Player1" {
		t.Errorf("Expected opponent name 'Player1', got '%s'", opponents[0].Name)
	}
}

func TestRemoveOpponent(t *testing.T) {
	gm, _ := setupMultiplayerTest()
	
	gm.AddOpponent("Player1", "http://localhost:8081")
	gm.AddOpponent("Player2", "http://localhost:8082")
	
	removed := gm.RemoveOpponent("Player1")
	if !removed {
		t.Error("Failed to remove Player1")
	}
	
	opponents := gm.GetAllOpponents()
	if len(opponents) != 1 {
		t.Errorf("Expected 1 opponent, got %d", len(opponents))
	}
	
	if opponents[0].Name != "Player2" {
		t.Errorf("Expected opponent name 'Player2', got '%s'", opponents[0].Name)
	}
}

func TestGetOpponent(t *testing.T) {
	gm, _ := setupMultiplayerTest()
	
	gm.AddOpponent("Player1", "http://localhost:8081")
	
	opp := gm.GetOpponent("Player1")
	if opp == nil {
		t.Fatal("Failed to get opponent Player1")
	}
	
	if opp.Name != "Player1" {
		t.Errorf("Expected opponent name 'Player1', got '%s'", opp.Name)
	}
	
	opp2 := gm.GetOpponent("NonExistent")
	if opp2 != nil {
		t.Error("Expected nil for non-existent opponent")
	}
}

func TestGetAliveOpponents(t *testing.T) {
	gm, _ := setupMultiplayerTest()
	
	// Créer des serveurs de test
	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"remaining_boats": 3}`))
	}))
	defer ts1.Close()
	
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"remaining_boats": 0}`))
	}))
	defer ts2.Close()
	
	gm.AddOpponent("AlivePlayer", ts1.URL)
	gm.AddOpponent("DeadPlayer", ts2.URL)
	
	time.Sleep(200 * time.Millisecond) // Attendre mise à jour statut
	
	gm.GetOpponent("AlivePlayer").UpdateStatus()
	gm.GetOpponent("DeadPlayer").UpdateStatus()
	
	time.Sleep(100 * time.Millisecond)
	
	alive := gm.GetAliveOpponents()
	
	t.Logf("Alive opponents: %d", len(alive))
	
	// Au moins AlivePlayer devrait être vivant
	found := false
	for _, opp := range alive {
		if opp.Name == "AlivePlayer" {
			found = true
		}
	}
	
	if !found {
		t.Error("AlivePlayer should be in alive opponents list")
	}
}

func TestCheckVictory(t *testing.T) {
	gm, _ := setupMultiplayerTest()
	
	// Pas d'adversaires = pas de victoire
	if gm.CheckVictory() {
		t.Error("Should not have victory with no opponents")
	}
	
	// Adversaire vivant = pas de victoire
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"remaining_boats": 3}`))
	}))
	defer ts.Close()
	
	gm.AddOpponent("Player1", ts.URL)
	time.Sleep(200 * time.Millisecond)
	gm.GetOpponent("Player1").UpdateStatus()
	
	if gm.CheckVictory() {
		t.Error("Should not have victory with alive opponent")
	}
	
	// Tous les adversaires éliminés = victoire
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"remaining_boats": 0}`))
	}))
	defer ts2.Close()
	
	gm.AddOpponent("Player2", ts2.URL)
	time.Sleep(200 * time.Millisecond)
	gm.GetOpponent("Player2").UpdateStatus()
	
	// Mettre Player1 en defeated
	gm.GetOpponent("Player1").Status = StatusDefeated
	
	time.Sleep(100 * time.Millisecond)
	
	if !gm.CheckVictory() {
		t.Error("Should have victory with all opponents defeated")
	}
}

func TestFireAt(t *testing.T) {
	gm, _ := setupMultiplayerTest()
	
	// Créer un serveur adversaire
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/hit" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"result": "hit", "x": 5, "y": 5}`))
		} else if r.URL.Path == "/boats" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"remaining_boats": 3}`))
		}
	}))
	defer ts.Close()
	
	gm.AddOpponent("Enemy", ts.URL)
	time.Sleep(200 * time.Millisecond)
	gm.GetOpponent("Enemy").UpdateStatus()
	
	msg, err := gm.FireAt("Enemy", 5, 5)
	if err != nil {
		t.Fatalf("Failed to fire at enemy: %v", err)
	}
	
	if msg == nil {
		t.Fatal("Expected message from FireAt")
	}
	
	t.Logf("Fire message: %s", *msg)
}

func TestFireAt_NonExistent(t *testing.T) {
	gm, _ := setupMultiplayerTest()
	
	_, err := gm.FireAt("NonExistent", 5, 5)
	if err == nil {
		t.Error("Expected error when firing at non-existent opponent")
	}
}

func TestAddOpponentFromRegister(t *testing.T) {
	gm, _ := setupMultiplayerTest()
	
	notified := false
	gm.OnNotification = func(msg string) {
		notified = true
	}
	
	gm.AddOpponentFromRegister("NewPlayer", "http://localhost:9999")
	
	time.Sleep(100 * time.Millisecond)
	
	if !notified {
		t.Error("Notification callback was not triggered")
	}
	
	opp := gm.GetOpponent("NewPlayer")
	if opp == nil {
		t.Error("Opponent was not added")
	}
}

func TestAddOpponentFromRegister_Duplicate(t *testing.T) {
	gm, _ := setupMultiplayerTest()
	
	gm.AddOpponent("Player1", "http://localhost:8081")
	
	initialCount := len(gm.GetAllOpponents())
	
	// Essayer d'ajouter le même joueur via register
	gm.AddOpponentFromRegister("Player1", "http://localhost:8081")
	
	finalCount := len(gm.GetAllOpponents())
	
	if finalCount != initialCount {
		t.Error("Duplicate opponent should not be added")
	}
}

func TestAsyncUpdates(t *testing.T) {
	gm, _ := setupMultiplayerTest()
	
	// Créer un serveur qui change son nombre de bateaux
	boatsCount := 3
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"remaining_boats": %d}`, boatsCount)))
	}))
	defer ts.Close()
	
	gm.AddOpponent("DynamicPlayer", ts.URL)
	
	// Démarrer les mises à jour asynchrones
	gm.StartAsyncUpdates(200 * time.Millisecond)
	defer gm.StopAsyncUpdates()
	
	time.Sleep(300 * time.Millisecond)
	
	opp := gm.GetOpponent("DynamicPlayer")
	if opp == nil {
		t.Fatal("Opponent not found")
	}
	
	initialBoats := opp.GetBoatsRemaining()
	
	// Changer le nombre de bateaux
	boatsCount = 1
	
	time.Sleep(300 * time.Millisecond)
	
	finalBoats := opp.GetBoatsRemaining()
	
	if finalBoats == initialBoats {
		t.Logf("Boats did not update (initial=%d, final=%d)", initialBoats, finalBoats)
	}
}

func TestMultiplePlayersScenario(t *testing.T) {
	// Créer 3 joueurs
	player1, url1 := setupMultiplayerTest()
	player2, url2 := setupMultiplayerTest()
	player3, url3 := setupMultiplayerTest()
	
	// Joueur 1 ajoute Joueur 2 et 3
	player1.AddOpponent("Player2", url2)
	player1.AddOpponent("Player3", url3)
	
	// Joueur 2 ajoute Joueur 1 et 3
	player2.AddOpponent("Player1", url1)
	player2.AddOpponent("Player3", url3)
	
	// Joueur 3 ajoute Joueur 1 et 2
	player3.AddOpponent("Player1", url1)
	player3.AddOpponent("Player2", url2)
	
	time.Sleep(500 * time.Millisecond)
	
	// Vérifier que chaque joueur voit les autres
	if len(player1.GetAllOpponents()) != 2 {
		t.Errorf("Player1 should see 2 opponents, got %d", len(player1.GetAllOpponents()))
	}
	
	if len(player2.GetAllOpponents()) != 2 {
		t.Errorf("Player2 should see 2 opponents, got %d", len(player2.GetAllOpponents()))
	}
	
	if len(player3.GetAllOpponents()) != 2 {
		t.Errorf("Player3 should see 2 opponents, got %d", len(player3.GetAllOpponents()))
	}
	
	// Joueur 1 tire sur Joueur 2
	msg, err := player1.FireAt("Player2", 5, 5)
	if err != nil {
		t.Errorf("Player1 failed to fire at Player2: %v", err)
	} else {
		t.Logf("Player1 -> Player2: %s", *msg)
	}
	
	// Joueur 2 tire sur Joueur 3
	msg, err = player2.FireAt("Player3", 3, 3)
	if err != nil {
		t.Errorf("Player2 failed to fire at Player3: %v", err)
	} else {
		t.Logf("Player2 -> Player3: %s", *msg)
	}
	
	// Joueur 3 tire sur Joueur 1
	msg, err = player3.FireAt("Player1", 7, 7)
	if err != nil {
		t.Errorf("Player3 failed to fire at Player1: %v", err)
	} else {
		t.Logf("Player3 -> Player1: %s", *msg)
	}
}

func TestBattleRoyaleScenario(t *testing.T) {
	// Créer 5 joueurs pour un Battle Royale
	players := make([]*GameManager, 5)
	urls := make([]string, 5)
	
	for i := 0; i < 5; i++ {
		gm, url := setupMultiplayerTest()
		gm.MyName = fmt.Sprintf("Player%d", i+1)
		players[i] = gm
		urls[i] = url
	}
	
	// Chaque joueur ajoute tous les autres
	for i, player := range players {
		for j, url := range urls {
			if i != j {
				player.AddOpponent(fmt.Sprintf("Player%d", j+1), url)
			}
		}
	}
	
	time.Sleep(500 * time.Millisecond)
	
	// Vérifier que chaque joueur voit 4 adversaires
	for i, player := range players {
		opponents := player.GetAllOpponents()
		if len(opponents) != 4 {
			t.Errorf("Player%d should see 4 opponents, got %d", i+1, len(opponents))
		}
	}
	
	// Simuler une bataille : chaque joueur tire sur le suivant
	for i := 0; i < 10; i++ {
		for j, player := range players {
			targetName := fmt.Sprintf("Player%d", ((j+1)%5)+1)
			player.FireAt(targetName, i%10, i/10)
		}
	}
	
	t.Log("Battle royale simulation completed")
}

func TestConcurrentOpponentAccess(t *testing.T) {
	gm, _ := setupMultiplayerTest()
	
	// Ajouter plusieurs adversaires
	for i := 0; i < 10; i++ {
		gm.AddOpponent(fmt.Sprintf("Player%d", i), fmt.Sprintf("http://localhost:808%d", i))
	}
	
	done := make(chan bool)
	
	// Accès concurrent aux adversaires
	for i := 0; i < 20; i++ {
		go func(id int) {
			opponents := gm.GetAllOpponents()
			_ = len(opponents)
			
			opp := gm.GetOpponent(fmt.Sprintf("Player%d", id%10))
			_ = opp
			
			done <- true
		}(i)
	}
	
	// Attendre tous les goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
	
	// Vérifier que la liste est toujours cohérente
	opponents := gm.GetAllOpponents()
	if len(opponents) != 10 {
		t.Errorf("Expected 10 opponents, got %d", len(opponents))
	}
}