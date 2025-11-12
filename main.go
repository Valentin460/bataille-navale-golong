package main

import (
	"bataille-navale/client"
	"bataille-navale/game"
	"bataille-navale/server"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	DefaultBoardSize = 10
	DefaultPort      = 8080
)

var (
	port      = flag.Int("port", DefaultPort, "Port du serveur HTTP")
	opponents = flag.String("opponents", "", "Liste des adresses des adversaires (séparées par des virgules)")
)

func main() {
	flag.Parse()
	
	boatSizes := []int{5, 4, 3, 3, 2}
	
	g := game.NewGame(DefaultBoardSize, boatSizes)
	
	fmt.Printf("=== BATAILLE NAVALE ===\n")
	fmt.Printf("Plateau initialisé: %dx%d\n", DefaultBoardSize, DefaultBoardSize)
	fmt.Printf("Nombre de bateaux: %d\n", len(boatSizes))
	fmt.Printf("Serveur démarré sur le port: %d\n", *port)
	fmt.Println()
	
	go startServer(g, *port)
	
	var clients []*client.Client
	if *opponents != "" {
		opponentsList := strings.Split(*opponents, ",")
		for _, addr := range opponentsList {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				c := client.NewClient(addr)
				clients = append(clients, c)
				fmt.Printf("Connecté à l'adversaire: %s\n", addr)
			}
		}
	}
	
	fmt.Println("\n=== COMMANDES DISPONIBLES ===")
	fmt.Println("- Interface interactive pour tirer sur les adversaires")
	fmt.Println("- Visualisation des plateaux adverses")
	fmt.Println("- Affichage de l'état de la partie")
	fmt.Println("\nLe serveur est actif. Utilisez Ctrl+C pour quitter.")
	
	select {}
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
