package cli

import (
	"bataille-navale/multiplayer"
	"fmt"
	"strings"
)

func DisplayOpponents(opponents []*multiplayer.Opponent) {
	if len(opponents) == 0 {
		fmt.Println(SymbolError + " Aucun adversaire connecté")
		return
	}
	
	fmt.Println("\n┌────────────────────────────────────────┐")
	fmt.Println("│         ADVERSAIRES CONNECTÉS          │")
	fmt.Println("├────────────────────────────────────────┤")
	
	for i, opp := range opponents {
		status := opp.StatusString()
		symbol := getStatusSymbol(opp.GetStatus())
		fmt.Printf("│ %d. %s %s%s│\n", 
			i+1, 
			symbol,
			padRight(opp.Name+" - "+status, 35),
		)
	}
	fmt.Println("└────────────────────────────────────────┘")
}

func DisplayBoard(board [][]int, title string) {
	fmt.Printf("\n%s\n", title)
	fmt.Println(strings.Repeat("─", 44))
	
	fmt.Print("   ")
	for x := 0; x < len(board[0]); x++ {
		fmt.Printf(" %d ", x)
	}
	fmt.Println()
	
	for y := 0; y < len(board); y++ {
		fmt.Printf("%2d ", y)
		for x := 0; x < len(board[y]); x++ {
			symbol := getCellSymbol(board[y][x])
			fmt.Printf(" %s ", symbol)
		}
		fmt.Println()
	}
}

func DisplayGameStatus(gm *multiplayer.GameManager) {
	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Println("║                    ÉTAT DU JEU                         ║")
	fmt.Println("╠════════════════════════════════════════════════════════╣")
	
	myStatusSymbol := ColorGreen + SymbolAlive + ColorReset
	myStatus := "Vivant"
	if !gm.AmIAlive() {
		myStatusSymbol = ColorRed + SymbolDefeated + ColorReset
		myStatus = "Éliminé"
	}
	fmt.Printf("║ Moi: %s %s%s║\n", myStatusSymbol, padRight(myStatus, 41))
	fmt.Printf("║ Mes bateaux: %s%s║\n", padRight(fmt.Sprintf("%d", gm.GetMyBoatsRemaining()), 39))
	fmt.Printf("║ Tirs reçus: %s%s║\n", padRight(fmt.Sprintf("%d", gm.GetReceivedHitsCount()), 40))
	
	fmt.Println("╠════════════════════════════════════════════════════════╣")
	
	opponents := gm.GetAllOpponents()
	if len(opponents) > 0 {
		aliveCount := 0
		for _, opp := range opponents {
			if opp.IsAlive() {
				aliveCount++
			}
		}
		fmt.Printf("║ Adversaires: %d vivants / %d total%s║\n", 
			aliveCount, len(opponents), padRight("", 23))
	} else {
		fmt.Println("║ Aucun adversaire                                       ║")
	}
	
	if gm.CheckVictory() {
		fmt.Println("╠════════════════════════════════════════════════════════╣")
		fmt.Println("║           *** VICTOIRE! TOUS ÉLIMINÉS! ***             ║")
	}
	
	fmt.Println("╚════════════════════════════════════════════════════════╝")
}

func DisplayMenu() {
	fmt.Println("\n┌──────────────── COMMANDES ───────────────┐")
	fmt.Println("│ fire [opp] [x] [y] - Tirer sur adversaire  │")
	fmt.Println("│ list               - Liste des adversaires │")
	fmt.Println("│ status             - État du jeu           │")
	fmt.Println("│ board [opp]        - Voir plateau adverse  │")
	fmt.Println("│ myboard            - Voir mon plateau      │")
	fmt.Println("│ add [nom] [url]    - Ajouter adversaire    │")
	fmt.Println("│ help               - Afficher l'aide       │")
	fmt.Println("│ quit               - Quitter               │")
	fmt.Println("└────────────────────────────────────────────┘")
}

// Helper functions
func getStatusSymbol(status multiplayer.OpponentStatus) string {
	switch status {
	case multiplayer.StatusAlive:
		return ColorGreen + SymbolAlive + ColorReset
	case multiplayer.StatusDefeated:
		return ColorRed + SymbolDefeated + ColorReset
	case multiplayer.StatusUnreachable:
		return ColorYellow + SymbolUnreachable + ColorReset
	default:
		return SymbolUnknown
	}
}

func getCellSymbol(state int) string {
	return FormatCell(state, true) // Avec couleurs
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
}

func ShowWelcome(port int) {
	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Println("║          BATAILLE NAVALE - JEU EN TEMPS RÉEL           ║")
	fmt.Println("╠════════════════════════════════════════════════════════╣")
	fmt.Printf("║ Serveur HTTP actif sur le port: %d%s║\n", port, padRight("", 18))
	fmt.Println("║ Votre plateau a été généré aléatoirement               ║")
	fmt.Println("║ Vous pouvez tirer sur vos adversaires à tout moment    ║")
	fmt.Println("║ Tapez 'help' pour voir les commandes disponibles       ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
}

func ShowNotification(msg string, msgType string) {
	symbol := FormatNotification(msgType, true)
	fmt.Printf("\n%s %s\n", symbol, msg)
}
