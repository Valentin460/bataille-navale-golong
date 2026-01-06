package models

type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

type Direction int

const (
	North Direction = iota // Haut
	South                  // Bas
	East                   // Droite
	West                   // Gauche
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

// TryMove tente de déplacer le bateau dans une direction
// Retourne les nouvelles coordonnées (x, y) si le mouvement est valide
func (b *Boat) TryMove(direction Direction, boardSize int) (newX, newY int, valid bool) {
	newX = b.X
	newY = b.Y
	
	switch direction {
	case North:
		newY = b.Y - 1
	case South:
		newY = b.Y + 1
	case East:
		newX = b.X + 1
	case West:
		newX = b.X - 1
	}
	
	// Vérifier que le bateau reste dans les limites
	if b.Orientation == Horizontal {
		if newX < 0 || newX+b.Size > boardSize || newY < 0 || newY >= boardSize {
			return b.X, b.Y, false
		}
	} else {
		if newX < 0 || newX >= boardSize || newY < 0 || newY+b.Size > boardSize {
			return b.X, b.Y, false
		}
	}
	
	return newX, newY, true
}

func (b *Boat) Move(newX, newY int) {
	b.X = newX
	b.Y = newY
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}
