// Package game implements a data representation for the game state of a Generals.io game session.
//
// This package can respond to update events to update itss internal model and provide helper
// accessors for the data. This version is heavily based on gioframework's implmentation
package game

import (
	"encoding/json"
)

// Game is a struct containing all of the gamestate of a Generals.io game session.
type Game struct {
	// c  *client.Client
	ID string

	chatroom string
	replayID string

	PreStart func()
	Start    func(playerindex int, users []string)
	Update   func(update gameUpdate)
	Won      func()
	Lost     func()
	Chat     func(user int, message string)

	lastAttack  int
	attackIndex int

	PlayerIndex int
	Width       int
	Height      int
	GameMap     []Cell
	inited      bool
	TurnCount   int

	mapRaw    []int
	citiesRaw []int

	Scores []struct {
		Armies int  `json:"total"`
		Tiles  int  `json:"tiles"`
		Index  int  `json:"i"`
		Dead   bool `json:"dead"`
	}
}

type gameUpdate struct {
	AttackIndex int   `json:"attackIndex"`
	CitiesDiff  []int `json:"cities_diff"`
	Generals    []int `json:"generals"`
	MapDiff     []int `json:"map_diff"`
	Scores      []struct {
		Armies int  `json:"total"`
		Tiles  int  `json:"tiles"`
		Index  int  `json:"i"`
		Dead   bool `json:"dead"`
	}
	Stars *[]float64 `json:"stars"`
	Turn  int        `json:"turn"`
}

// PreGameStart empty
func (g *Game) PreGameStart() {}

// GameStart process game init
func (g *Game) GameStart(raw json.RawMessage) {
	gameinfo := struct {
		PlayerIndex int      `json:"playerIndex"`
		ReplayID    string   `json:"replay_id"`
		ChatRoom    string   `json:"chat_room"`
		Usernames   []string `json:"usernames"`
	}{}
	decode := []interface{}{nil, &gameinfo}
	json.Unmarshal(raw, &decode)
	g.PlayerIndex = gameinfo.PlayerIndex
	g.chatroom = gameinfo.ChatRoom
	g.replayID = gameinfo.ReplayID
	if g.Start != nil {
		g.Start(gameinfo.PlayerIndex, gameinfo.Usernames)
	}
}

// GameUpdate process game update
func (g *Game) GameUpdate(raw json.RawMessage) {
	update := gameUpdate{}
	decode := []interface{}{nil, &update}
	json.Unmarshal(raw, &decode)

	newRaw := []int{}
	difPos := 0
	oldPos := 0
	for difPos < len(update.MapDiff) {
		getOld := update.MapDiff[difPos]
		difPos++
		for l1 := 0; l1 < getOld; l1++ {
			newRaw = append(newRaw, g.mapRaw[oldPos])
			oldPos++
		}
		if difPos >= len(update.MapDiff) {
			break
		}
		getNew := update.MapDiff[difPos]
		difPos++
		for l1 := 0; l1 < getNew; l1++ {
			newRaw = append(newRaw, update.MapDiff[difPos])
			oldPos++
			difPos++
		}
	}

	g.mapRaw = newRaw

	newRaw = []int{}
	difPos = 0
	oldPos = 0
	for difPos < len(update.CitiesDiff) {
		getOld := update.CitiesDiff[difPos]
		difPos++
		for l1 := 0; l1 < getOld; l1++ {
			newRaw = append(newRaw, g.citiesRaw[oldPos])
			oldPos++
		}
		if difPos >= len(update.CitiesDiff) {
			break
		}
		getNew := update.CitiesDiff[difPos]
		difPos++
		for l1 := 0; l1 < getNew; l1++ {
			newRaw = append(newRaw, update.CitiesDiff[difPos])
			oldPos++
			difPos++
		}
	}
	g.citiesRaw = newRaw

	if !g.inited {
		g.Width = g.mapRaw[0]
		g.Height = g.mapRaw[1]
		g.GameMap = make([]Cell, g.Width*g.Height)
		g.inited = true
	}

	g.TurnCount = update.Turn
	g.attackIndex = update.AttackIndex

	g.Scores = update.Scores

	for x := 0; x < g.Width; x++ {
		for y := 0; y < g.Height; y++ {
			g.GameMap[y*g.Width+x].Armies = g.mapRaw[y*g.Width+x+2]
		}
	}
	for x := 0; x < g.Width; x++ {
		for y := 0; y < g.Height; y++ {
			g.GameMap[y*g.Width+x].Faction = g.mapRaw[y*g.Width+x+2+g.Width*g.Height]
		}
	}
	for _, city := range g.citiesRaw {
		if city >= 0 {
			g.GameMap[city].Type = City
		}
	}
	for _, general := range update.Generals {
		if general >= 0 {
			g.GameMap[general].Type = General
		}
	}

	if g.Update != nil {
		g.Update(update)
	}
}

// GameWon empty
func (g *Game) GameWon() {}

// GameLost empty
func (g *Game) GameLost() {}

// GameOver empty
func (g *Game) GameOver() {}

// Dispatch handles any game events
func (g *Game) Dispatch(event string, data json.RawMessage) {
	switch event {
	case "game_start":

	case "game_update":

	}
}

// GetAdjacents returns a list of map indicies of cells adgacent to the given cell
func (g *Game) GetAdjacents(from int) (adjacent []int) {
	if from >= g.Width {
		adjacent = append(adjacent, from-g.Width)
	}
	if from < g.Width*(g.Height-1) {
		adjacent = append(adjacent, from+g.Width)
	}
	if from%g.Width > 0 {
		adjacent = append(adjacent, from-1)
	}
	if from%g.Width < g.Width-1 {
		adjacent = append(adjacent, from+1)
	}
	return
}

// GetNeighborhood seems to return a 3x3 area of map indicies around the given cell?
func (g *Game) GetNeighborhood(from int) (adjacent []int) {
	if from >= g.Width {
		if from%g.Width > 0 {
			adjacent = append(adjacent, (from-g.Width)-1)
		}
		adjacent = append(adjacent, from-g.Width)
		if from%g.Width < g.Width-1 {
			adjacent = append(adjacent, (from-g.Width)+1)
		}
	}
	if from < g.Width*(g.Height-1) {
		if from%g.Width > 0 {
			adjacent = append(adjacent, (from+g.Width)-1)
		}
		adjacent = append(adjacent, from+g.Width)
		if from%g.Width < g.Width-1 {
			adjacent = append(adjacent, (from+g.Width)+1)
		}
	}
	if from%g.Width > 0 {
		adjacent = append(adjacent, from-1)
	}
	if from%g.Width < g.Width-1 {
		adjacent = append(adjacent, from+1)
	}
	return
}

// GetDistance returns the Manhatten distance between two map indicies
func (g *Game) GetDistance(from, to int) int {

	x1, y1 := from%g.Width, from/g.Width
	x2, y2 := to%g.Width, to/g.Width
	dx := x1 - x2
	dy := y1 - y2
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// SendChat sends a message to the current chatroom
// func (g *Game) SendChat(msg string) {
// 	g.c.sendMessage("chat_message", g.chatroom, msg)
// }

// QueueLength is how many attacks we have queued up
func (g *Game) QueueLength() int {
	return g.lastAttack - g.attackIndex
}

// Walkable is true if the given cell is not a mountain or fog obstacle
func (g *Game) Walkable(cell int) bool {
	return g.GameMap[cell].Faction != -2 && g.GameMap[cell].Faction != -4
}

// NextAttackIndex temp function meant to simulate the old behavior of attack indexes
func (g *Game) NextAttackIndex() int {
	g.lastAttack++
	return g.lastAttack
}

// Attack sends an attack request to the server
// func (g *Game) Attack(from, to int, is50 bool) {
// 	g.lastAttack++
// 	g.c.Attack(from, to, is50, g.lastAttack)
// }
