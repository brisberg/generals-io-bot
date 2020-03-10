package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/brisberg/generals-io-bot/client"
)

func main() {
	fmt.Printf("Starting Generals AI Program:\n")

	client, err := client.Connect("bot")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close("Main Program Terminated")

	client.OnClose = func() {
		log.Println("Close Callback. Exiting.")
		os.Exit(1)
	}

	go client.Run()

	if err := client.RegisterBot("mybot-batz", "[Bot]Keidence-45"); err != nil {
		log.Fatalln(err)
	}
	time.Sleep(2000 * time.Millisecond)

	client.JoinCustomGame("botbotbot")
	time.Sleep(3000 * time.Millisecond)

	client.SetForceStart(true)
	time.Sleep(3000 * time.Millisecond)

	for {
		time.Sleep(100 * time.Millisecond)
		if client.Game.QueueLength() > 0 {
			continue
		}
		mine := []int{}
		for i, tile := range client.Game.GameMap {
			if tile.Faction == client.Game.PlayerIndex && tile.Armies > 1 {
				mine = append(mine, i)
			}
		}
		if len(mine) == 0 {
			continue
		}
		cell := rand.Intn(len(mine))
		move := []int{}
		for _, adjacent := range client.Game.GetAdjacents(mine[cell]) {
			if client.Game.Walkable(adjacent) {
				move = append(move, adjacent)
			}
		}
		if len(move) == 0 {
			continue
		}
		movecell := rand.Intn(len(move))
		client.Attack(mine[cell], move[movecell], false)
	}

	// client.LeaveLobby()
	// time.Sleep(10000 * time.Millisecond)
}
