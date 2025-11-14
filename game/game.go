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
		placed := false
		attempts := 0
		maxAttempts := 1000
		
		for !placed && attempts < maxAttempts {
			attempts++
			
			orientation := models.Horizontal
			if rand.Intn(2) == 1 {
				orientation = models.Vertical
			}
			
			var x, y int
			if orientation == models.Horizontal {
				x = rand.Intn(g.Board.Size - size + 1)
				y = rand.Intn(g.Board.Size)
			} else {
				x = rand.Intn(g.Board.Size)
				y = rand.Intn(g.Board.Size - size + 1)
			}
			
			if g.canPlaceBoat(x, y, size, orientation) {
				boat := &models.Boat{
					ID:          i,
					Size:        size,
					X:           x,
					Y:           y,
					Orientation: orientation,
					HitCount:    0,
				}
				g.Boats = append(g.Boats, boat)
				
				g.markBoatOnBoard(boat)
				placed = true
			}
		}
		
		if !placed {
			g.placeBoatAnywhere(i, size)
		}
	}
}

func (g *Game) canPlaceBoat(x, y, size int, orientation models.Orientation) bool {
	for i := 0; i < size; i++ {
		checkX := x
		checkY := y
		
		if orientation == models.Horizontal {
			checkX = x + i
		} else {
			checkY = y + i
		}
		
		if !g.Board.IsValidPosition(checkX, checkY) {
			return false
		}
		
		if g.Board.Cells[checkY][checkX].HasBoat {
			return false
		}
	}
	return true
}

func (g *Game) markBoatOnBoard(boat *models.Boat) {
	positions := boat.GetPositions()
	for _, pos := range positions {
		g.Board.Cells[pos.Y][pos.X].HasBoat = true
		g.Board.Cells[pos.Y][pos.X].BoatID = boat.ID
	}
}

func (g *Game) placeBoatAnywhere(id, size int) {
	for y := 0; y < g.Board.Size; y++ {
		for x := 0; x < g.Board.Size; x++ {
			if g.canPlaceBoat(x, y, size, models.Horizontal) {
				boat := &models.Boat{
					ID:          id,
					Size:        size,
					X:           x,
					Y:           y,
					Orientation: models.Horizontal,
					HitCount:    0,
				}
				g.Boats = append(g.Boats, boat)
				g.markBoatOnBoard(boat)
				return
			}
			
			if g.canPlaceBoat(x, y, size, models.Vertical) {
				boat := &models.Boat{
					ID:          id,
					Size:        size,
					X:           x,
					Y:           y,
					Orientation: models.Vertical,
					HitCount:    0,
				}
				g.Boats = append(g.Boats, boat)
				g.markBoatOnBoard(boat)
				return
			}
		}
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
