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

	// c.UseGameConstructor(func() client.IGame {
	// 	return &game.Game{}
	// })

	go c.Run()

	if err := c.RegisterBot("mybot-batz", "[Bot]Keidence-45"); err != nil {
		log.Fatalln(err)
	}
	time.Sleep(2000 * time.Millisecond)

	// Launch a go routine for playing the game. Listens on the game events queue
	// var g *game.Game
	g := &game.Game{}
	gameEvents := make(chan client.NetworkEvent)
	c.SetGameEventChan(gameEvents)
	go func(gameEvents <-chan client.NetworkEvent, g *game.Game) {
		for evt := range gameEvents {
			switch evt.Name {
			case "pre_game_start":
				// g = &game.Game{}
				g.PreGameStart()
			case "game_start":
				g.GameStart(evt.Data)
			case "game_update":
				g.GameUpdate(evt.Data)
			case "game_won":
				g.GameWon()
			case "game_lost":
				g.GameLost()
			case "game_over":
				g.GameOver()
				// g = nil
			}
		}
	}(gameEvents, g)

	c.JoinCustomGame("botbotbot")
	time.Sleep(3000 * time.Millisecond)

	c.SetForceStart(true)
	time.Sleep(3000 * time.Millisecond)

	for {
		time.Sleep(5000 * time.Millisecond)
		if g != nil {
			log.Println("Game has started, starting bot...")
			for {
				time.Sleep(100 * time.Millisecond)
				if g.QueueLength() > 0 {
					continue
				}
				mine := []int{}
				for i, tile := range g.GameMap {
					if tile.Faction == g.PlayerIndex && tile.Armies > 1 {
						mine = append(mine, i)
					}
				}
				if len(mine) == 0 {
					continue
				}
				cell := rand.Intn(len(mine))
				move := []int{}
				for _, adjacent := range g.GetAdjacents(mine[cell]) {
					if g.Walkable(adjacent) {
						move = append(move, adjacent)
					}
				}
				if len(move) == 0 {
					continue
				}
				movecell := rand.Intn(len(move))
				c.Attack(mine[cell], move[movecell], false, g.NextAttackIndex())
			}
		}
	}

	// c.LeaveLobby()
	// time.Sleep(10000 * time.Millisecond)
}
