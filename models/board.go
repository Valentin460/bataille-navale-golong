package models

type CellState int

const (
	Empty CellState = iota
	Miss
	Hit
)

type Cell struct {
	State    CellState
	HasBoat  bool
	BoatID   int
	Revealed bool
}

type Board struct {
	Size  int
	Cells [][]Cell
}

func NewBoard(size int) *Board {
	cells := make([][]Cell, size)
	for i := range cells {
		cells[i] = make([]Cell, size)
	}
	return &Board{
		Size:  size,
		Cells: cells,
	}
}

func (b *Board) IsValidPosition(x, y int) bool {
	return x >= 0 && x < b.Size && y >= 0 && y < b.Size
}
