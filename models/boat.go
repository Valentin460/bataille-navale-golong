package models

type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

type Boat struct {
	ID          int
	Size        int
	X           int
	Y           int
	Orientation Orientation
	HitCount    int
}

func (b *Boat) IsAfloat() bool {
	return b.HitCount < b.Size
}

func (b *Boat) GetPositions() []Position {
	positions := make([]Position, b.Size)
	for i := 0; i < b.Size; i++ {
		if b.Orientation == Horizontal {
			positions[i] = Position{X: b.X + i, Y: b.Y}
		} else {
			positions[i] = Position{X: b.X, Y: b.Y + i}
		}
	}
	return positions
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}
