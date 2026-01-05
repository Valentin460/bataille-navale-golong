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
	
	go startServer(g, *port)
	
	gm := multiplayer.NewGameManager(g)
	
	if *opponents != "" {
		opponentsList := strings.Split(*opponents, ",")
		for i, addr := range opponentsList {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				name := fmt.Sprintf("player%d", i+1)
				gm.AddOpponent(name, addr)
			}
		}
	}
	
	cli.ShowWelcome(*port)
	
	interactiveCLI := cli.NewInteractiveCLI(gm)
	interactiveCLI.Start()
}

func startServer(g *game.Game, port int) {
	s := server.NewServer(g)
	mux := http.NewServeMux()
	s.SetupRoutes(mux)
	
	addr := fmt.Sprintf(":%d", port)
	
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
		os.Exit(1)
	}
}
