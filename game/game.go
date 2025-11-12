package game

import (
	"bataille-navale/models"
	"math/rand"
	"sync"
	"time"
)

type Game struct {
	Board         *models.Board
	Boats         []*models.Boat
	ReceivedHits  []models.HitInfo
	mu            sync.RWMutex
}

func NewGame(boardSize int, boatSizes []int) *Game {
	rand.Seed(time.Now().UnixNano())
	
	game := &Game{
		Board:        models.NewBoard(boardSize),
		Boats:        make([]*models.Boat, 0),
		ReceivedHits: make([]models.HitInfo, 0),
	}
	
	game.placeBoatsRandomly(boatSizes)
	
	return game
}

func (g *Game) placeBoatsRandomly(sizes []int) {
	for i, size := range sizes {
		boat := &models.Boat{
			ID:          i,
			Size:        size,
			X:           0,
			Y:           0,
			Orientation: models.Horizontal,
			HitCount:    0,
		}
		g.Boats = append(g.Boats, boat)
	}
}

func (g *Game) ProcessHit(x, y int) models.HitResponse {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	if !g.Board.IsValidPosition(x, y) {
		return models.HitResponse{
			Result: "invalid",
			X:      x,
			Y:      y,
		}
	}
	
	cell := &g.Board.Cells[y][x]
	cell.Revealed = true
	
	result := "miss"
	if cell.HasBoat {
		cell.State = models.Hit
		result = "hit"
		
		for _, boat := range g.Boats {
			if boat.ID == cell.BoatID {
				boat.HitCount++
				break
			}
		}
	} else {
		cell.State = models.Miss
	}
	
	g.ReceivedHits = append(g.ReceivedHits, models.HitInfo{
		X:      x,
		Y:      y,
		Result: result,
	})
	
	return models.HitResponse{
		Result: result,
		X:      x,
		Y:      y,
	}
}

func (g *Game) GetRemainingBoats() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	count := 0
	for _, boat := range g.Boats {
		if boat.IsAfloat() {
			count++
		}
	}
	return count
}

func (g *Game) GetBoardState() models.BoardResponse {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	cells := make([][]int, g.Board.Size)
	for i := range cells {
		cells[i] = make([]int, g.Board.Size)
		for j := range cells[i] {
			cell := g.Board.Cells[i][j]
			if !cell.Revealed {
				cells[i][j] = 0
			} else if cell.State == models.Miss {
				cells[i][j] = 1
			} else if cell.State == models.Hit {
				cells[i][j] = 2
			}
		}
	}
	
	return models.BoardResponse{
		Size:  g.Board.Size,
		Cells: cells,
	}
}

func (g *Game) GetReceivedHits() models.HitsResponse {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	return models.HitsResponse{
		Hits: g.ReceivedHits,
	}
}

func (g *Game) IsAlive() bool {
	return g.GetRemainingBoats() > 0
}
