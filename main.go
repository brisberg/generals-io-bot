package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/brisberg/generals-io-bot/client"
	"github.com/brisberg/generals-io-bot/game"
)

func main() {
	fmt.Printf("Starting Generals AI Program:\n")

	c, err := client.Connect("bot")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close("Main Program Terminated")

	c.OnClose = func() {
		log.Println("Close Callback. Exiting.")
		os.Exit(1)
	}

	c.UseGameConstructor(func() client.IGame {
		return &game.Game{}
	})

	go c.Run()

	if err := c.RegisterBot("mybot-batz", "[Bot]Keidence-45"); err != nil {
		log.Fatalln(err)
	}
	time.Sleep(2000 * time.Millisecond)

	c.JoinCustomGame("botbotbot")
	time.Sleep(3000 * time.Millisecond)

	c.SetForceStart(true)
	time.Sleep(3000 * time.Millisecond)

	game := game.Game(c.Game)
	for {
		time.Sleep(100 * time.Millisecond)
		if game.QueueLength() > 0 {
			continue
		}
		mine := []int{}
		for i, tile := range game.GameMap {
			if tile.Faction == game.PlayerIndex && tile.Armies > 1 {
				mine = append(mine, i)
			}
		}
		if len(mine) == 0 {
			continue
		}
		cell := rand.Intn(len(mine))
		move := []int{}
		for _, adjacent := range game.GetAdjacents(mine[cell]) {
			if game.Walkable(adjacent) {
				move = append(move, adjacent)
			}
		}
		if len(move) == 0 {
			continue
		}
		movecell := rand.Intn(len(move))
		c.Attack(mine[cell], move[movecell], false)
	}

	// c.LeaveLobby()
	// time.Sleep(10000 * time.Millisecond)
}
