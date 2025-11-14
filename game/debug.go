package game

import (
	"fmt"
)

func (g *Game) DebugPrintBoard() {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	fmt.Println("\n=== Plateau de jeu ===")
	fmt.Print("   ")
	for x := 0; x < g.Board.Size; x++ {
		fmt.Printf(" %d ", x)
	}
	fmt.Println()
	
	for y := 0; y < g.Board.Size; y++ {
		fmt.Printf("%2d ", y)
		for x := 0; x < g.Board.Size; x++ {
			cell := g.Board.Cells[y][x]
			if cell.HasBoat {
				if cell.Revealed {
					if cell.State == 2 {
						fmt.Print(" X ")
					} else {
						fmt.Print(" B ")
					}
				} else {
					fmt.Print(" B ")
				}
			} else {
				if cell.Revealed && cell.State == 1 {
					fmt.Print(" o ")
				} else {
					fmt.Print(" ~ ")
				}
			}
		}
		fmt.Println()
	}
	
	fmt.Println("\n=== Bateaux placés ===")
	for _, boat := range g.Boats {
		orientation := "Horizontal"
		if boat.Orientation == 1 {
			orientation = "Vertical"
		}
		fmt.Printf("Bateau %d: Taille=%d, Position=(%d,%d), Orientation=%s, Touches=%d/%d\n",
			boat.ID, boat.Size, boat.X, boat.Y, orientation, boat.HitCount, boat.Size)
	}
	fmt.Println()
}
