package client

import (
	"bataille-navale/models"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// UI gère l'interface utilisateur du client
type UI struct {
	LocalClient     *Client
	OpponentClients []*OpponentClient
	scanner         *bufio.Scanner
}

// OpponentClient représente un adversaire avec son client et ses données
type OpponentClient struct {
	Client   *Client
	Address  string
	Board    [][]int // Plateau vu de notre côté (0=inconnu, 1=touché, -1=raté)
	IsAlive  bool
	Name     string
	LastHits []models.HitInfo
}

// NewUI crée une nouvelle interface utilisateur
func NewUI(localAddr string) *UI {
	return &UI{
		LocalClient:     NewClient(localAddr),
		OpponentClients: make([]*OpponentClient, 0),
		scanner:         bufio.NewScanner(os.Stdin),
	}
}

// AddOpponent ajoute un adversaire
func (ui *UI) AddOpponent(addr string) {
	opponent := &OpponentClient{
		Client:  NewClient(addr),
		Address: addr,
		Board:   make([][]int, 10),
		IsAlive: true,
		Name:    fmt.Sprintf("Adversaire %d", len(ui.OpponentClients)+1),
	}

	// Initialiser le plateau de l'adversaire
	for i := range opponent.Board {
		opponent.Board[i] = make([]int, 10)
	}

	ui.OpponentClients = append(ui.OpponentClients, opponent)
}

// ClearScreen efface l'écran du terminal
func (ui *UI) ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// DisplayBoard affiche le plateau personnel
func (ui *UI) DisplayBoard() error {
	board, err := ui.LocalClient.GetBoard()
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération du plateau: %w", err)
	}

	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║        VOTRE PLATEAU PERSONNEL         ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	// En-tête avec les colonnes
	fmt.Print("    ")
	for i := 0; i < board.Size; i++ {
		fmt.Printf(" %d ", i)
	}
	fmt.Println()

	fmt.Print("   ╔")
	for i := 0; i < board.Size; i++ {
		fmt.Print("═══")
	}
	fmt.Println("╗")

	// Affichage des lignes
	for y := 0; y < board.Size; y++ {
		fmt.Printf(" %d ║", y)
		for x := 0; x < board.Size; x++ {
			cell := board.Cells[y][x]
			switch cell {
			case 0: // Vide
				fmt.Print(" . ")
			case 1: // Bateau
				fmt.Print(" B ")
			case 2: // Raté
				fmt.Print(" ○ ")
			case 3: // Touché
				fmt.Print(" X ")
			default:
				fmt.Print(" ? ")
			}
		}
		fmt.Println("║")
	}

	fmt.Print("   ╚")
	for i := 0; i < board.Size; i++ {
		fmt.Print("═══")
	}
	fmt.Println("╝")

	return nil
}

// DisplayOpponentBoard affiche le plateau d'un adversaire
func (ui *UI) DisplayOpponentBoard(opponent *OpponentClient) {
	fmt.Printf("\n╔════════════════════════════════════════╗\n")
	fmt.Printf("║  %s %-26s ║\n", opponent.Name, fmt.Sprintf("(%s)", opponent.Address))
	if !opponent.IsAlive {
		fmt.Printf("║           [ÉLIMINÉ]                    ║\n")
	}
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	// En-tête avec les colonnes
	fmt.Print("    ")
	for i := 0; i < len(opponent.Board); i++ {
		fmt.Printf(" %d ", i)
	}
	fmt.Println()

	fmt.Print("   ╔")
	for i := 0; i < len(opponent.Board); i++ {
		fmt.Print("═══")
	}
	fmt.Println("╗")

	// Affichage des lignes
	for y := 0; y < len(opponent.Board); y++ {
		fmt.Printf(" %d ║", y)
		for x := 0; x < len(opponent.Board[y]); x++ {
			cell := opponent.Board[y][x]
			switch cell {
			case 0: // Inconnu
				fmt.Print(" . ")
			case 1: // Touché
				fmt.Print(" X ")
			case -1: // Raté
				fmt.Print(" ○ ")
			default:
				fmt.Print(" ? ")
			}
		}
		fmt.Println("║")
	}

	fmt.Print("   ╚")
	for i := 0; i < len(opponent.Board); i++ {
		fmt.Print("═══")
	}
	fmt.Println("╝")
}

// DisplayGameStats affiche les statistiques de la partie
func (ui *UI) DisplayGameStats() error {
	// Récupérer les informations locales
	boatsCount, err := ui.LocalClient.GetBoatsCount()
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération des bateaux: %w", err)
	}

	hits, err := ui.LocalClient.GetHits()
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération des tirs: %w", err)
	}

	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║        STATISTIQUES DE JEU             ║")
	fmt.Println("╠════════════════════════════════════════╣")

	// Bateaux restants
	fmt.Printf("║ Bateaux restants : %-19d ║\n", boatsCount)

	// Tirs reçus
	receivedHits := 0
	impacts := 0
	for _, hit := range hits.Hits {
		receivedHits++
		if hit.Result == "touché" || hit.Result == "coulé" {
			impacts++
		}
	}

	fmt.Printf("║ Tirs reçus       : %-19d ║\n", receivedHits)
	fmt.Printf("║ Impacts subis    : %-19d ║\n", impacts)

	// État du joueur
	if boatsCount == 0 {
		fmt.Println("║                                        ║")
		fmt.Println("║         🚫 VOUS ÊTES ÉLIMINÉ 🚫        ║")
	}

	fmt.Println("╚════════════════════════════════════════╝")

	// Derniers tirs reçus
	if len(hits.Hits) > 0 {
		fmt.Println("\n┌────────────────────────────────────────┐")
		fmt.Println("│      DERNIERS TIRS REÇUS               │")
		fmt.Println("├────────────────────────────────────────┤")

		// Afficher les 5 derniers tirs
		start := 0
		if len(hits.Hits) > 5 {
			start = len(hits.Hits) - 5
		}

		for i := start; i < len(hits.Hits); i++ {
			hit := hits.Hits[i]
			status := "○"
			if hit.Result == "touché" {
				status = "X"
			} else if hit.Result == "coulé" {
				status = "⚓"
			}
			fmt.Printf("│ %s (%d,%d) : %-25s │\n", status, hit.X, hit.Y, hit.Result)
		}

		fmt.Println("└────────────────────────────────────────┘")
	}

	return nil
}

// DisplayOpponentStats affiche les statistiques des adversaires
func (ui *UI) DisplayOpponentStats() {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║         ÉTAT DES ADVERSAIRES           ║")
	fmt.Println("╠════════════════════════════════════════╣")

	for i, opp := range ui.OpponentClients {
		status := "✓ Actif"
		if !opp.IsAlive {
			status = "✗ Éliminé"
		}

		fmt.Printf("║ %d. %-25s %s ║\n", i+1, opp.Name, status)
	}

	fmt.Println("╚════════════════════════════════════════╝")
}

// DisplayMainScreen affiche l'écran principal
func (ui *UI) DisplayMainScreen() error {
	ui.ClearScreen()

	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              🚢 BATAILLE NAVALE - CLIENT 🚢                    ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")

	// Afficher le plateau personnel
	if err := ui.DisplayBoard(); err != nil {
		return err
	}

	// Afficher les statistiques
	if err := ui.DisplayGameStats(); err != nil {
		return err
	}

	// Afficher l'état des adversaires
	if len(ui.OpponentClients) > 0 {
		ui.DisplayOpponentStats()
	}

	return nil
}

// ShootAtOpponent effectue un tir sur un adversaire
func (ui *UI) ShootAtOpponent(opponentIndex, x, y int) error {
	if opponentIndex < 0 || opponentIndex >= len(ui.OpponentClients) {
		return fmt.Errorf("adversaire invalide")
	}

	opponent := ui.OpponentClients[opponentIndex]

	if !opponent.IsAlive {
		return fmt.Errorf("cet adversaire est déjà éliminé")
	}

	// Effectuer le tir
	result, err := opponent.Client.Hit(x, y)
	if err != nil {
		return fmt.Errorf("erreur lors du tir: %w", err)
	}

	// Mettre à jour le plateau de l'adversaire
	if result.Result == "touché" || result.Result == "coulé" {
		opponent.Board[y][x] = 1 // Touché
	} else {
		opponent.Board[y][x] = -1 // Raté
	}

	// Vérifier si l'adversaire est toujours en vie
	if !opponent.Client.IsAlive() {
		opponent.IsAlive = false
	}

	return nil
}

// ReadInput lit une entrée de l'utilisateur
func (ui *UI) ReadInput(prompt string) string {
	fmt.Print(prompt)
	ui.scanner.Scan()
	return strings.TrimSpace(ui.scanner.Text())
}

// InteractiveShoot gère la saisie interactive des tirs
func (ui *UI) InteractiveShoot() error {
	// Vérifier si le joueur est éliminé
	boats, err := ui.LocalClient.GetBoatsCount()
	if err != nil {
		return err
	}

	if boats == 0 {
		fmt.Println("\n❌ Vous êtes éliminé ! Vous ne pouvez plus tirer.")
		ui.ReadInput("\nAppuyez sur Entrée pour continuer...")
		return nil
	}

	if len(ui.OpponentClients) == 0 {
		fmt.Println("\n⚠️  Aucun adversaire configuré.")
		ui.ReadInput("\nAppuyez sur Entrée pour continuer...")
		return nil
	}

	// Sélectionner un adversaire
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║         SÉLECTION DE L'ADVERSAIRE      ║")
	fmt.Println("╚════════════════════════════════════════╝")

	for i, opp := range ui.OpponentClients {
		status := "[Actif]"
		if !opp.IsAlive {
			status = "[Éliminé]"
		}
		fmt.Printf("%d. %s - %s %s\n", i+1, opp.Name, opp.Address, status)
	}

	opponentInput := ui.ReadInput("\nChoisissez un adversaire (numéro) : ")
	opponentIndex, err := strconv.Atoi(opponentInput)
	if err != nil || opponentIndex < 1 || opponentIndex > len(ui.OpponentClients) {
		fmt.Println("❌ Adversaire invalide.")
		ui.ReadInput("\nAppuyez sur Entrée pour continuer...")
		return nil
	}

	opponentIndex-- // Convertir en index 0-based
	opponent := ui.OpponentClients[opponentIndex]

	// Afficher le plateau de l'adversaire
	ui.ClearScreen()
	ui.DisplayOpponentBoard(opponent)

	// Saisir les coordonnées
	fmt.Println("\n📍 Entrez les coordonnées du tir :")
	xInput := ui.ReadInput("  X (0-9) : ")
	yInput := ui.ReadInput("  Y (0-9) : ")

	x, errX := strconv.Atoi(xInput)
	y, errY := strconv.Atoi(yInput)

	if errX != nil || errY != nil || x < 0 || x > 9 || y < 0 || y > 9 {
		fmt.Println("❌ Coordonnées invalides.")
		ui.ReadInput("\nAppuyez sur Entrée pour continuer...")
		return nil
	}

	// Effectuer le tir
	fmt.Printf("\n🎯 Tir en cours sur (%d,%d)...\n", x, y)

	if err := ui.ShootAtOpponent(opponentIndex, x, y); err != nil {
		fmt.Printf("❌ Erreur : %v\n", err)
	} else {
		result := opponent.Board[y][x]
		if result == 1 {
			fmt.Println("💥 TOUCHÉ !")
		} else {
			fmt.Println("💨 RATÉ !")
		}

		if !opponent.IsAlive {
			fmt.Printf("🎊 %s a été éliminé !\n", opponent.Name)
		}
	}

	ui.ReadInput("\nAppuyez sur Entrée pour continuer...")
	return nil
}

// Run démarre l'interface interactive
func (ui *UI) Run() {
	for {
		if err := ui.DisplayMainScreen(); err != nil {
			fmt.Printf("Erreur: %v\n", err)
		}

		fmt.Println("\n╔════════════════════════════════════════╗")
		fmt.Println("║              MENU PRINCIPAL            ║")
		fmt.Println("╠════════════════════════════════════════╣")
		fmt.Println("║ 1. Tirer sur un adversaire             ║")
		fmt.Println("║ 2. Voir les plateaux adverses          ║")
		fmt.Println("║ 3. Actualiser                          ║")
		fmt.Println("║ 4. Quitter                             ║")
		fmt.Println("╚════════════════════════════════════════╝")

		choice := ui.ReadInput("\nVotre choix : ")

		switch choice {
		case "1":
			ui.InteractiveShoot()
		case "2":
			ui.ShowAllOpponentBoards()
		case "3":
			// Actualiser (boucle)
		case "4":
			fmt.Println("\n👋 Au revoir !")
			os.Exit(0)
		default:
			fmt.Println("❌ Choix invalide.")
			ui.ReadInput("\nAppuyez sur Entrée pour continuer...")
		}
	}
}

// ShowAllOpponentBoards affiche tous les plateaux adverses
func (ui *UI) ShowAllOpponentBoards() {
	if len(ui.OpponentClients) == 0 {
		fmt.Println("\n⚠️  Aucun adversaire configuré.")
		ui.ReadInput("\nAppuyez sur Entrée pour continuer...")
		return
	}

	ui.ClearScreen()
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    PLATEAUX ADVERSES                           ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")

	for _, opp := range ui.OpponentClients {
		ui.DisplayOpponentBoard(opp)
		fmt.Println()
	}

	ui.ReadInput("\nAppuyez sur Entrée pour continuer...")
}
