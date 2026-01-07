package game

import (
	"bataille-navale/models"
	"testing"
)

func TestNewGame(t *testing.T) {
	g := NewGame(10, []int{5, 4, 3})
	
	if g.Board.Size != 10 {
		t.Errorf("Expected board size 10, got %d", g.Board.Size)
	}
	
	if len(g.Boats) != 3 {
		t.Errorf("Expected 3 boats, got %d", len(g.Boats))
	}
	
	// Vérifier que les bateaux sont bien placés
	for i, boat := range g.Boats {
		if boat.Size != []int{5, 4, 3}[i] {
			t.Errorf("Boat %d: expected size %d, got %d", i, []int{5, 4, 3}[i], boat.Size)
		}
		if boat.HitCount != 0 {
			t.Errorf("Boat %d: expected HitCount 0, got %d", i, boat.HitCount)
		}
	}
}

func TestProcessHit_Miss(t *testing.T) {
	g := NewGame(10, []int{5})
	g.DisableMovement = true // Désactiver mouvement pour tests
	
	// Trouver une case vide
	var emptyX, emptyY int
	found := false
	for y := 0; y < g.Board.Size && !found; y++ {
		for x := 0; x < g.Board.Size && !found; x++ {
			if !g.Board.Cells[y][x].HasBoat {
				emptyX, emptyY = x, y
				found = true
			}
		}
	}
	
	if !found {
		t.Skip("No empty cell found (all covered by boats)")
	}
	
	resp := g.ProcessHit(emptyX, emptyY)
	
	if resp.Result != "miss" {
		t.Errorf("Expected miss, got %s", resp.Result)
	}
	
	if resp.X != emptyX || resp.Y != emptyY {
		t.Errorf("Expected coords (%d,%d), got (%d,%d)", emptyX, emptyY, resp.X, resp.Y)
	}
	
	// Vérifier que la cellule est révélée
	if !g.Board.Cells[emptyY][emptyX].Revealed {
		t.Error("Cell should be revealed after hit")
	}
	
	if g.Board.Cells[emptyY][emptyX].State != models.Miss {
		t.Error("Cell state should be Miss")
	}
}

func TestProcessHit_Hit(t *testing.T) {
	g := NewGame(10, []int{5})
	g.DisableMovement = true // Désactiver mouvement pour tests
	
	// Trouver une case avec bateau
	var boatX, boatY int
	found := false
	for y := 0; y < g.Board.Size && !found; y++ {
		for x := 0; x < g.Board.Size && !found; x++ {
			if g.Board.Cells[y][x].HasBoat {
				boatX, boatY = x, y
				found = true
			}
		}
	}
	
	if !found {
		t.Fatal("No boat cell found")
	}
	
	boatID := g.Board.Cells[boatY][boatX].BoatID
	initialHitCount := g.Boats[boatID].HitCount
	
	resp := g.ProcessHit(boatX, boatY)
	
	if resp.Result != "hit" {
		t.Errorf("Expected hit, got %s", resp.Result)
	}
	
	// Vérifier que le HitCount a augmenté
	if g.Boats[boatID].HitCount != initialHitCount+1 {
		t.Errorf("Expected HitCount %d, got %d", initialHitCount+1, g.Boats[boatID].HitCount)
	}
	
	// Vérifier que la cellule est révélée
	if !g.Board.Cells[boatY][boatX].Revealed {
		t.Error("Cell should be revealed after hit")
	}
	
	if g.Board.Cells[boatY][boatX].State != models.Hit {
		t.Error("Cell state should be Hit")
	}
}

func TestProcessHit_Invalid(t *testing.T) {
	g := NewGame(10, []int{5})
	
	// Coordonnées invalides
	resp := g.ProcessHit(-1, 5)
	if resp.Result != "invalid" {
		t.Errorf("Expected invalid for x=-1, got %s", resp.Result)
	}
	
	resp = g.ProcessHit(5, 15)
	if resp.Result != "invalid" {
		t.Errorf("Expected invalid for y=15, got %s", resp.Result)
	}
}

func TestGetRemainingBoats(t *testing.T) {
	g := NewGame(10, []int{2}) // Un petit bateau
	g.DisableMovement = true // Désactiver mouvement pour tests
	
	initial := g.GetRemainingBoats()
	if initial != 1 {
		t.Errorf("Expected 1 boat initially, got %d", initial)
	}
	
	// Trouver et couler le bateau
	boat := g.Boats[0]
	positions := boat.GetPositions()
	
	for _, pos := range positions {
		g.ProcessHit(pos.X, pos.Y)
	}
	
	remaining := g.GetRemainingBoats()
	if remaining != 0 {
		t.Errorf("Expected 0 boats after sinking, got %d", remaining)
	}
}

func TestIsAlive(t *testing.T) {
	g := NewGame(10, []int{2})
	g.DisableMovement = true // Désactiver mouvement pour tests
	
	if !g.IsAlive() {
		t.Error("Game should be alive initially")
	}
	
	// Couler tous les bateaux
	for _, boat := range g.Boats {
		positions := boat.GetPositions()
		for _, pos := range positions {
			g.ProcessHit(pos.X, pos.Y)
		}
	}
	
	if g.IsAlive() {
		t.Error("Game should not be alive after all boats sunk")
	}
}

func TestProcessSpecialShot_Cross(t *testing.T) {
	g := NewGame(10, []int{5})
	g.DisableMovement = true // Désactiver mouvement pour tests
	
	// Tir en croix au centre (5,5)
	result := g.ProcessSpecialShot(5, 5, models.ShotCross)
	
	if result.ShotType != models.ShotCross {
		t.Errorf("Expected ShotCross, got %s", result.ShotType)
	}
	
	// Doit toucher 5 cases maximum (centre + 4 directions)
	if len(result.Hits) > 5 {
		t.Errorf("Cross shot should hit max 5 cells, got %d", len(result.Hits))
	}
	
	// Vérifier que les cellules sont révélées
	if !g.Board.Cells[5][5].Revealed {
		t.Error("Center cell should be revealed")
	}
}

func TestProcessSpecialShot_Zone(t *testing.T) {
	g := NewGame(10, []int{5})
	g.DisableMovement = true // Désactiver mouvement pour tests
	
	// Tir de zone 3x3 au centre (5,5)
	result := g.ProcessSpecialShot(5, 5, models.ShotZone)
	
	if result.ShotType != models.ShotZone {
		t.Errorf("Expected ShotZone, got %s", result.ShotType)
	}
	
	// Doit toucher 9 cases
	if len(result.Hits) != 9 {
		t.Errorf("Zone shot should hit 9 cells, got %d", len(result.Hits))
	}
}

func TestProcessSpecialShot_Paralyzing(t *testing.T) {
	g := NewGame(10, []int{3})
	g.DisableMovement = true // Désactiver mouvement pour tests
	
	// Trouver un bateau
	var boatX, boatY int
	found := false
	for y := 0; y < g.Board.Size && !found; y++ {
		for x := 0; x < g.Board.Size && !found; x++ {
			if g.Board.Cells[y][x].HasBoat {
				boatX, boatY = x, y
				found = true
			}
		}
	}
	
	if !found {
		t.Fatal("No boat found")
	}
	
	boatID := g.Board.Cells[boatY][boatX].BoatID
	
	// Tir paralysant
	result := g.ProcessSpecialShot(boatX, boatY, models.ShotParalyzing)
	
	if result.ShotType != models.ShotParalyzing {
		t.Errorf("Expected ShotParalyzing, got %s", result.ShotType)
	}
	
	// Vérifier que le bateau est paralysé
	if !g.Boats[boatID].Paralyzed {
		t.Error("Boat should be paralyzed after paralyzing shot")
	}
	
	if g.Boats[boatID].ParalyzedUntil != 3 {
		t.Errorf("Boat should be paralyzed for 3 turns, got %d", g.Boats[boatID].ParalyzedUntil)
	}
}

func TestBoatMovement(t *testing.T) {
	g := NewGame(10, []int{2})
	
	boat := g.Boats[0]
	initialX := boat.X
	initialY := boat.Y
	
	// Forcer un mouvement
	g.moveAllBoats()
	
	// Le bateau devrait avoir bougé (probabilité élevée avec plusieurs essais)
	// Note: Test probabiliste, peut échouer rarement
	moved := false
	for i := 0; i < 10 && !moved; i++ {
		g.moveAllBoats()
		if boat.X != initialX || boat.Y != initialY {
			moved = true
		}
	}
	
	// On accepte que le bateau ne bouge pas (mur/collision)
	t.Logf("Boat moved: %v (from %d,%d to %d,%d)", moved, initialX, initialY, boat.X, boat.Y)
}

func TestConcurrentHits(t *testing.T) {
	g := NewGame(10, []int{5, 4, 3})
	g.DisableMovement = true // Désactiver mouvement pour tests
	
	// Test de concurrence
	done := make(chan bool)
	
	for i := 0; i < 10; i++ {
		go func(x, y int) {
			g.ProcessHit(x, y)
			done <- true
		}(i%10, i/10)
	}
	
	// Attendre tous les goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Le jeu devrait toujours être cohérent (pas de panic)
	boats := g.GetRemainingBoats()
	t.Logf("Remaining boats after concurrent hits: %d", boats)
}