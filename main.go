package main

import (
	"bataille-navale/cli"
	"bataille-navale/game"
	"bataille-navale/multiplayer"
	"bataille-navale/server"
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
	port      = flag.Int("port", DefaultPort, "Port du serveur HTTP")
	opponents = flag.String("opponents", "", "Liste des adresses des adversaires (séparées par des virgules)")
	name      = flag.String("name", "", "Nom du joueur (optionnel, généré automatiquement si vide)")
)

func main() {
	flag.Parse()
	
	// Générer un nom unique si non fourni
	playerName := *name
	if playerName == "" {
		playerName = fmt.Sprintf("player%d", time.Now().Unix()%10000)
	}
	
	myURL := fmt.Sprintf("http://localhost:%d", *port)
	
	boatSizes := []int{5, 4, 3, 3, 2}
	
	g := game.NewGame(DefaultBoardSize, boatSizes)
	
	gm := multiplayer.NewGameManager(g, playerName, myURL)
	
	// Démarrer le serveur HTTP
	s := server.NewServer(g)
	
	// Connecter le callback pour recevoir les enregistrements mutuels
	s.OnOpponentAdded = func(oppName, oppURL string) {
		gm.AddOpponentFromRegister(oppName, oppURL)
	}
	
	go startServer(s, *port)
	
	// Ajouter les adversaires de la ligne de commande
	if *opponents != "" {
		opponentsList := strings.Split(*opponents, ",")
		for i, addr := range opponentsList {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				oppName := fmt.Sprintf("player%d", i+1)
				gm.AddOpponent(oppName, addr)
			}
		}
	}
	
	cli.ShowWelcome(*port)
	fmt.Printf("👤 Votre nom de joueur: %s\n", playerName)
	
	interactiveCLI := cli.NewInteractiveCLI(gm)
	
	// Connecter le callback pour afficher les notifications
	gm.OnNotification = func(msg string) {
		fmt.Println("\n" + msg)
		fmt.Print("\n> ") // Re-afficher le prompt
	}
	
	interactiveCLI.Start()
}

func startServer(s *server.Server, port int) {
	mux := http.NewServeMux()
	s.SetupRoutes(mux)
	
	addr := fmt.Sprintf(":%d", port)
	
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
		os.Exit(1)
	}
}