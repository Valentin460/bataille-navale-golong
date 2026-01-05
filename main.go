package main

import (
	"bataille-navale/client"
	"bataille-navale/game"
	"bataille-navale/server"
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	DefaultBoardSize = 10
	DefaultPort      = 8080
)

var (
	port       = flag.Int("port", DefaultPort, "Port du serveur HTTP")
	opponents  = flag.String("opponents", "", "Liste des adresses des adversaires (séparées par des virgules)")
	clientMode = flag.Bool("client", false, "Démarrer en mode client interactif")
)

func main() {
	flag.Parse()

	boatSizes := []int{5, 4, 3, 3, 2}

	g := game.NewGame(DefaultBoardSize, boatSizes)

	fmt.Printf("╔════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              🚢 BATAILLE NAVALE - GOLONG 🚢                    ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════╝\n")
	fmt.Printf("\n📊 Plateau initialisé: %dx%d\n", DefaultBoardSize, DefaultBoardSize)
	fmt.Printf("⚓ Nombre de bateaux: %d\n", len(boatSizes))
	fmt.Printf("🌐 Serveur démarré sur le port: %d\n", *port)
	fmt.Println()

	// Démarrer le serveur dans une goroutine
	go startServer(g, *port)

	// Attendre que le serveur soit prêt
	time.Sleep(500 * time.Millisecond)

	// Construire l'adresse locale
	localAddr := fmt.Sprintf("http://localhost:%d", *port)

	// Créer l'interface utilisateur
	ui := client.NewUI(localAddr)

	// Récupérer les adresses adverses
	opponentAddrs := getOpponentAddresses(*opponents)

	// Ajouter les adversaires
	for _, addr := range opponentAddrs {
		ui.AddOpponent(addr)
		fmt.Printf("✓ Adversaire ajouté: %s\n", addr)
	}

	if len(opponentAddrs) == 0 {
		fmt.Println("\n⚠️  Aucun adversaire configuré.")
		fmt.Println("💡 Vous pouvez ajouter des adversaires avec --opponents=http://addr1,http://addr2")
		fmt.Println("   ou les entrer manuellement ci-dessous (Entrée vide pour terminer)")

		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("\n🎯 Adresse adversaire : ")
			scanner.Scan()
			addr := strings.TrimSpace(scanner.Text())

			if addr == "" {
				break
			}

			// Ajouter http:// si absent
			if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
				addr = "http://" + addr
			}

			ui.AddOpponent(addr)
			fmt.Printf("✓ Adversaire ajouté: %s\n", addr)
		}
	}

	fmt.Println("\n🎮 Démarrage de l'interface interactive...")
	time.Sleep(1 * time.Second)

	// Lancer l'interface interactive
	ui.Run()
}

func getOpponentAddresses(opponentsFlag string) []string {
	var addresses []string

	if opponentsFlag != "" {
		parts := strings.Split(opponentsFlag, ",")
		for _, addr := range parts {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				// Ajouter http:// si absent
				if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
					addr = "http://" + addr
				}
				addresses = append(addresses, addr)
			}
		}
	}

	return addresses
}

func startServer(g *game.Game, port int) {
	s := server.NewServer(g)
	mux := http.NewServeMux()
	s.SetupRoutes(mux)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Serveur en écoute sur http://localhost%s\n", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
		os.Exit(1)
	}
}
