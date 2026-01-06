package models

// ShotType définit le type de tir
type ShotType string

const (
	ShotNormal     ShotType = "normal"     // Tir normal (1 case)
	ShotCross      ShotType = "cross"      // Tir en croix (5 cases)
	ShotZone       ShotType = "zone"       // Tir de zone (9 cases 3x3)
	ShotParalyzing ShotType = "paralyzing" // Tir paralysant (1 case, empêche mouvement)
)

// ShotRequest représente une requête de tir avec type
type ShotRequest struct {
	X        int      `json:"x"`
	Y        int      `json:"y"`
	ShotType ShotType `json:"shot_type,omitempty"`
}

// ShotResult représente le résultat d'un tir (peut toucher plusieurs cases)
type ShotResult struct {
	Hits      []HitResponse `json:"hits"`       // Liste des impacts
	ShotType  ShotType      `json:"shot_type"`  // Type de tir utilisé
	TotalHits int           `json:"total_hits"` // Nombre de touches
}

func (s ShotType) GetTargetPositions(x, y int, boardSize int) []Position {
	positions := []Position{}

	switch s {
	case ShotNormal, ShotParalyzing:
		// Tir normal ou paralysant : 1 case
		if x >= 0 && x < boardSize && y >= 0 && y < boardSize {
			positions = append(positions, Position{X: x, Y: y})
		}

	case ShotCross:
		// Tir en croix : centre + 4 directions
		// Centre
		if x >= 0 && x < boardSize && y >= 0 && y < boardSize {
			positions = append(positions, Position{X: x, Y: y})
		}
		// Nord
		if x >= 0 && x < boardSize && y-1 >= 0 && y-1 < boardSize {
			positions = append(positions, Position{X: x, Y: y - 1})
		}
		// Sud
		if x >= 0 && x < boardSize && y+1 >= 0 && y+1 < boardSize {
			positions = append(positions, Position{X: x, Y: y + 1})
		}
		// Est
		if x+1 >= 0 && x+1 < boardSize && y >= 0 && y < boardSize {
			positions = append(positions, Position{X: x + 1, Y: y})
		}
		// Ouest
		if x-1 >= 0 && x-1 < boardSize && y >= 0 && y < boardSize {
			positions = append(positions, Position{X: x - 1, Y: y})
		}

	case ShotZone:
		// Tir de zone : 3x3 autour du point
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				nx := x + dx
				ny := y + dy
				if nx >= 0 && nx < boardSize && ny >= 0 && ny < boardSize {
					positions = append(positions, Position{X: nx, Y: ny})
				}
			}
		}
	}

	return positions
}

func (s ShotType) Description() string {
	switch s {
	case ShotNormal:
		return "Tir normal (1 case)"
	case ShotCross:
		return "Tir en croix (5 cases: centre + directions)"
	case ShotZone:
		return "Tir de zone (9 cases: 3x3)"
	case ShotParalyzing:
		return "Tir paralysant (1 case, bloque mouvement)"
	default:
		return "Type de tir inconnu"
	}
}

func (s ShotType) IsValid() bool {
	switch s {
	case ShotNormal, ShotCross, ShotZone, ShotParalyzing:
		return true
	default:
		return false
	}
}
