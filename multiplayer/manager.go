package multiplayer

import (
	"bataille-navale/game"
	"fmt"
	"sync"
	"time"
)

type GameManager struct {
	MyGame         *game.Game
	MyName         string    // Nom du joueur local
	MyURL          string    // URL du joueur local
	Opponents      []*Opponent
	mu             sync.RWMutex
	stopUpdate     chan bool
	OnNotification func(message string) // Callback pour afficher des notifications dans le CLI
}

func NewGameManager(myGame *game.Game, myName, myURL string) *GameManager {
	return &GameManager{
		MyGame:     myGame,
		MyName:     myName,
		MyURL:      myURL,
		Opponents:  make([]*Opponent, 0),
		mu:         sync.RWMutex{},
		stopUpdate: make(chan bool),
	}
}

func (gm *GameManager) AddOpponent(name, url string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	
	opponent := NewOpponent(name, url)
	gm.Opponents = append(gm.Opponents, opponent)
	
	// Mise à jour initiale du statut
	go opponent.UpdateStatus()
	
	// Enregistrement mutuel : notifier l'adversaire qu'on l'a ajouté
	go func() {
		time.Sleep(500 * time.Millisecond) // Petit délai pour que le serveur soit prêt
		err := opponent.Client.Register(gm.MyName, gm.MyURL)
		if err != nil {
			// Enregistrement échoué, mais ce n'est pas critique
			fmt.Printf("⚠️  Échec de l'enregistrement mutuel avec %s: %v\n", name, err)
		}
	}()
}

// AddOpponentFromRegister ajoute un adversaire suite à une demande /register
func (gm *GameManager) AddOpponentFromRegister(name, url string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	
	// Vérifier si l'adversaire existe déjà
	for _, opp := range gm.Opponents {
		if opp.Name == name {
			return // Déjà présent
		}
	}
	
	opponent := NewOpponent(name, url)
	gm.Opponents = append(gm.Opponents, opponent)
	
	go opponent.UpdateStatus()
	
	// Notifier le CLI
	if gm.OnNotification != nil {
		msg := fmt.Sprintf("🔔 Joueur '%s' vous a ajouté! Tapez 'list' pour voir.", name)
		gm.OnNotification(msg)
	}
}

func (gm *GameManager) RemoveOpponent(name string) bool {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	
	for i, opp := range gm.Opponents {
		if opp.Name == name {
			gm.Opponents = append(gm.Opponents[:i], gm.Opponents[i+1:]...)
			return true
		}
	}
	return false
}

func (gm *GameManager) GetOpponent(name string) *Opponent {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	
	for _, opp := range gm.Opponents {
		if opp.Name == name {
			return opp
		}
	}
	return nil
}

func (gm *GameManager) GetAllOpponents() []*Opponent {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	
	opponents := make([]*Opponent, len(gm.Opponents))
	copy(opponents, gm.Opponents)
	return opponents
}

func (gm *GameManager) GetAliveOpponents() []*Opponent {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	
	alive := make([]*Opponent, 0)
	for _, opp := range gm.Opponents {
		if opp.IsAlive() {
			alive = append(alive, opp)
		}
	}
	return alive
}

func (gm *GameManager) StartAsyncUpdates(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				gm.updateAllOpponents()
			case <-gm.stopUpdate:
				return
			}
		}
	}()
}

func (gm *GameManager) StopAsyncUpdates() {
	gm.stopUpdate <- true
}

func (gm *GameManager) updateAllOpponents() {
	opponents := gm.GetAllOpponents()
	
	var wg sync.WaitGroup
	for _, opp := range opponents {
		wg.Add(1)
		go func(o *Opponent) {
			defer wg.Done()
			o.UpdateStatus()
		}(opp)
	}
	wg.Wait()
}

func (gm *GameManager) FireAt(opponentName string, x, y int) (*string, error) {
	opponent := gm.GetOpponent(opponentName)
	if opponent == nil {
		return nil, fmt.Errorf("adversaire %s introuvable", opponentName)
	}
	
	if !opponent.IsAlive() {
		msg := fmt.Sprintf("L'adversaire %s est éliminé ou injoignable", opponentName)
		return &msg, fmt.Errorf("%s", msg)
	}
	
	resp, err := opponent.Hit(x, y)
	if err != nil {
		msg := fmt.Sprintf("Erreur: %v", err)
		return &msg, err
	}
	
	var msg string
	if resp.Result == "hit" {
		msg = fmt.Sprintf("🎯 TOUCHÉ! (%d,%d) sur %s", x, y, opponentName)
	} else {
		msg = fmt.Sprintf("💧 Dans l'eau (%d,%d) sur %s", x, y, opponentName)
	}
	
	go opponent.UpdateStatus()
	
	return &msg, nil
}

func (gm *GameManager) AmIAlive() bool {
	return gm.MyGame.IsAlive()
}

func (gm *GameManager) GetMyBoatsRemaining() int {
	return gm.MyGame.GetRemainingBoats()
}

func (gm *GameManager) GetGameStatus() string {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	
	myStatus := "Vivant"
	if !gm.MyGame.IsAlive() {
		myStatus = "Éliminé"
	}
	
	aliveCount := 0
	defeatedCount := 0
	unreachableCount := 0
	
	for _, opp := range gm.Opponents {
		switch opp.GetStatus() {
		case StatusAlive, StatusUnknown:
			aliveCount++
		case StatusDefeated:
			defeatedCount++
		case StatusUnreachable:
			unreachableCount++
		}
	}
	
	return fmt.Sprintf(
		"Moi: %s (%d bateaux) | Adversaires: %d vivants, %d éliminés, %d injoignables",
		myStatus,
		gm.MyGame.GetRemainingBoats(),
		aliveCount,
		defeatedCount,
		unreachableCount,
	)
}

func (gm *GameManager) CheckVictory() bool {
	if !gm.AmIAlive() {
		return false
	}
	
	aliveOpponents := gm.GetAliveOpponents()
	return len(aliveOpponents) == 0 && len(gm.Opponents) > 0
}

func (gm *GameManager) GetReceivedHitsCount() int {
	hits := gm.MyGame.GetReceivedHits()
	return len(hits.Hits)
}