package cli

import (
	"bataille-navale/multiplayer"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type InteractiveCLI struct {
	GameManager *multiplayer.GameManager
	scanner     *bufio.Scanner
	running     bool
}

func NewInteractiveCLI(gm *multiplayer.GameManager) *InteractiveCLI {
	return &InteractiveCLI{
		GameManager: gm,
		scanner:     bufio.NewScanner(os.Stdin),
		running:     true,
	}
}

func (cli *InteractiveCLI) Start() {
	cli.GameManager.StartAsyncUpdates(3 * time.Second)
	
	DisplayMenu()
	
	for cli.running {
		fmt.Print("\n> ")
		if !cli.scanner.Scan() {
			break
		}
		
		input := strings.TrimSpace(cli.scanner.Text())
		if input == "" {
			continue
		}
		
		cli.handleCommand(input)
	}
	
	cli.GameManager.StopAsyncUpdates()
}

func (cli *InteractiveCLI) handleCommand(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}
	
	command := strings.ToLower(parts[0])
	
	switch command {
	case "fire":
		cli.commandFire(parts)
	case "list":
		cli.commandList()
	case "status":
		cli.commandStatus()
	case "board":
		cli.commandBoard(parts)
	case "myboard":
		cli.commandMyBoard()
	case "add":
		cli.commandAdd(parts)
	case "help":
		DisplayMenu()
	case "quit", "exit":
		cli.commandQuit()
	default:
		ShowNotification("Commande inconnue. Tapez 'help' pour l'aide.", "error")
	}
}

func (cli *InteractiveCLI) commandFire(parts []string) {
	if len(parts) < 4 {
		ShowNotification("Usage: fire [nom_adversaire] [x] [y]", "error")
		return
	}
	
	if !cli.GameManager.AmIAlive() {
		ShowNotification("Vous êtes éliminé! Impossible de tirer.", "error")
		return
	}
	
	opponentName := parts[1]
	x, err1 := strconv.Atoi(parts[2])
	y, err2 := strconv.Atoi(parts[3])
	
	if err1 != nil || err2 != nil {
		ShowNotification("Coordonnées invalides. Utilisez des nombres.", "error")
		return
	}
	
	if x < 0 || x > 9 || y < 0 || y > 9 {
		ShowNotification("Coordonnées hors limites (0-9).", "error")
		return
	}
	
	msg, err := cli.GameManager.FireAt(opponentName, x, y)
	if err != nil {
		ShowNotification(fmt.Sprintf("Erreur: %v", err), "error")
		return
	}
	
	if msg != nil {
		fmt.Printf("\n%s %s\n", SymbolHitShot, *msg)
	}
}

func (cli *InteractiveCLI) commandList() {
	opponents := cli.GameManager.GetAllOpponents()
	DisplayOpponents(opponents)
}

func (cli *InteractiveCLI) commandStatus() {
	DisplayGameStatus(cli.GameManager)
}

func (cli *InteractiveCLI) commandBoard(parts []string) {
	if len(parts) < 2 {
		ShowNotification("Usage: board [nom_adversaire]", "error")
		return
	}
	
	opponentName := parts[1]
	opponent := cli.GameManager.GetOpponent(opponentName)
	
	if opponent == nil {
		ShowNotification(fmt.Sprintf("Adversaire '%s' introuvable", opponentName), "error")
		return
	}
	
	board := opponent.GetKnownBoard()
	if len(board) == 0 || len(board[0]) == 0 {
		ShowNotification("Aucun tir effectué sur cet adversaire", "warning")
		return
	}
	
	title := fmt.Sprintf("PLATEAU DE %s (%s)", strings.ToUpper(opponentName), opponent.StatusString())
	DisplayBoard(board, title)
}

func (cli *InteractiveCLI) commandMyBoard() {
	state := cli.GameManager.MyGame.GetBoardState()
	
	board := make([][]int, state.Size)
	for i := range board {
		board[i] = make([]int, state.Size)
		copy(board[i], state.Cells[i])
	}
	
	for _, boat := range cli.GameManager.MyGame.Boats {
		positions := boat.GetPositions()
		for _, pos := range positions {
			if board[pos.Y][pos.X] == 0 {
				board[pos.Y][pos.X] = 3
			}
		}
	}
	
	fmt.Printf("\nMON PLATEAU (Bateaux: %d)\n", cli.GameManager.GetMyBoatsRemaining())
	fmt.Println(strings.Repeat("─", 44))
	
	// En-tête
	fmt.Print("   ")
	for x := 0; x < state.Size; x++ {
		fmt.Printf(" %d ", x)
	}
	fmt.Println()
	
	// Lignes
	for y := 0; y < state.Size; y++ {
		fmt.Printf("%2d ", y)
		for x := 0; x < state.Size; x++ {
			symbol := FormatCell(board[y][x], true)
			fmt.Printf(" %s ", symbol)
		}
		fmt.Println()
	}
	
	fmt.Printf("\nLégende: %s eau, %s bateau, %s tir raté, %s touché\n", 
		SymbolWater, SymbolBoat, SymbolMiss, SymbolHit)
}

func (cli *InteractiveCLI) commandAdd(parts []string) {
	if len(parts) < 3 {
		ShowNotification("Usage: add [nom] [url]", "error")
		fmt.Println("Exemple: add player1 http://localhost:8081")
		return
	}
	
	name := parts[1]
	url := parts[2]
	
	if cli.GameManager.GetOpponent(name) != nil {
		ShowNotification(fmt.Sprintf("Adversaire '%s' déjà ajouté", name), "warning")
		return
	}
	
	cli.GameManager.AddOpponent(name, url)
	ShowNotification(fmt.Sprintf("Adversaire '%s' ajouté! Connexion en cours...", name), "success")
	
	time.Sleep(1 * time.Second)
	opponent := cli.GameManager.GetOpponent(name)
	if opponent != nil {
		fmt.Printf("Statut: %s\n", opponent.StatusString())
	}
}

func (cli *InteractiveCLI) commandQuit() {
	ShowNotification("Au revoir!", "info")
	cli.running = false
}
