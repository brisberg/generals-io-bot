// Package client provides utilities for interacting with Generals.io over a WebSocket connection.
//
// Game adds types and interfaces for the Client to interact with a user provide Game module.
package client

import "encoding/json"

// BaseGame is a base type used by the client to store the gamestate of a game on Generals.io
// Specific bot/game implmentations should extend this type
type BaseGame struct {
}

// IGame interface is the common interface for all Game implementaions
// It can respond to all update events from the Client
type IGame interface {
	PreGameStart()
	GameStart(raw json.RawMessage)
	GameUpdate(raw json.RawMessage)
	GameWon()
	GameLost()
	GameOver()

	NextAttackIndex() int
}
