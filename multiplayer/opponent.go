package multiplayer

import (
	"bataille-navale/client"
	"bataille-navale/models"
	"fmt"
	"sync"
	"time"
)

type OpponentStatus int

const (
	StatusUnknown OpponentStatus = iota
	StatusAlive
	StatusDefeated
	StatusUnreachable
)

type Opponent struct {
	Name          string
	Client        *client.Client
	Status        OpponentStatus
	LastCheck     time.Time
	BoatsRemaining int
	KnownBoard    [][]int
	mu            sync.RWMutex
}

func NewOpponent(name, url string) *Opponent {
	return &Opponent{
		Name:       name,
		Client:     client.NewClient(url),
		Status:     StatusUnknown,
		KnownBoard: make([][]int, 10),
		mu:         sync.RWMutex{},
	}
}

func (o *Opponent) UpdateStatus() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	boats, err := o.Client.GetBoatsCount()
	if err != nil {
		o.Status = StatusUnreachable
		return err
	}

	o.BoatsRemaining = boats
	o.LastCheck = time.Now()

	if boats == 0 {
		o.Status = StatusDefeated
	} else {
		o.Status = StatusAlive
	}

	return nil
}

func (o *Opponent) Hit(x, y int) (*models.HitResponse, error) {
	resp, err := o.Client.Hit(x, y)
	if err != nil {
		o.mu.Lock()
		o.Status = StatusUnreachable
		o.mu.Unlock()
		return nil, err
	}

	o.mu.Lock()
	if len(o.KnownBoard) == 0 || len(o.KnownBoard[0]) == 0 {
		o.KnownBoard = make([][]int, 10)
		for i := range o.KnownBoard {
			o.KnownBoard[i] = make([]int, 10)
		}
	}
	
	if resp.Result == "hit" {
		o.KnownBoard[y][x] = 2
	} else if resp.Result == "miss" {
		o.KnownBoard[y][x] = 1
	}
	o.mu.Unlock()

	return resp, nil
}

func (o *Opponent) GetStatus() OpponentStatus {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.Status
}

func (o *Opponent) IsAlive() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.Status == StatusAlive || o.Status == StatusUnknown
}

func (o *Opponent) GetBoatsRemaining() int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.BoatsRemaining
}

func (o *Opponent) GetKnownBoard() [][]int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	board := make([][]int, len(o.KnownBoard))
	for i := range o.KnownBoard {
		board[i] = make([]int, len(o.KnownBoard[i]))
		copy(board[i], o.KnownBoard[i])
	}
	return board
}

func (o *Opponent) StatusString() string {
	switch o.GetStatus() {
	case StatusAlive:
		return fmt.Sprintf("Vivant (%d bateaux)", o.GetBoatsRemaining())
	case StatusDefeated:
		return "Éliminé"
	case StatusUnreachable:
		return "Injoignable"
	default:
		return "Inconnu"
	}
}
