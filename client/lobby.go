// Package client provides utilities for interacting with Generals.io over a WebSocket connection.
//
// Lobby adds types and utility functions for joining/leaving and interacting with Lobbies on Generals.io.
package client

import (
	"errors"
	"log"
)

// Lobby is a type representing a game lobby on Generals.io
type Lobby struct {
	// Client used to interact with this lobby (might remove this?)
	c *Client
	// Name of the custom lobby
	name string
	// Custom Map title if lobby is using one
	mapTitle string
	// True if we are voting to force a start
	isForcing bool
	// Number of players voting to force a start
	numForce int
	// Index of current player in the lobby
	lobbyIndex int
	// Array of player indicies
	playerIndices []int
	// Usernames of all players (in index order)
	usernames []string
	// Team ID of all players (in index order)
	teams []int
}

// NewLobby create a new Lobby instance with the given name
func NewLobby(name string) *Lobby {
	return &Lobby{
		name: name,
	}
}

// QueueUpdate JSON Datastructure returned by server for most state changes in a Lobby
type QueueUpdate struct {
	MapTitle      string   `json:"mapTitle"`
	LobbyIndex    int      `json:"lobbyIndex"`
	IsForcing     bool     `json:"isForcing"`
	NumForce      int      `json:"numForce"`
	PlayerIndices []int    `json:"playerIndices"`
	Usernames     []string `json:"usernames"`
	Teams         []int    `json:"teams"`
	// Options TODO: Implement tracking options
}

// Update updates Lobby instance with a information from server
func (l *Lobby) Update(update QueueUpdate) {
	l.lobbyIndex = update.LobbyIndex
	l.mapTitle = update.MapTitle
	l.isForcing = update.IsForcing
	l.numForce = update.NumForce
	l.playerIndices = update.PlayerIndices
	l.usernames = update.Usernames
	l.teams = update.Teams
}

// JoinCustomGame joins a custom game with the specified ID. Doesn't return the game object
func (c *Client) JoinCustomGame(ID string) {
	log.Println("Joined custom game at http://bot.generals.io/games/", ID)
	c.sendMessage(msg, "join_private", ID, c.userID)
	c.lobby = NewLobby(ID)
	// g := &Game{c: c, ID: ID}
	// g.registerEvents()
	// return g
}

// SetForceStart changes our force start status in the current Lobby
func (c *Client) SetForceStart(force bool) error {
	if c.lobby == nil {
		return errors.New("Error: Can't force start a game when not in a Lobby. Try joining a game first")
	}

	c.sendMessage(msg, "set_force_start", c.lobby.name, force)
	return nil
}

// LeaveLobby leaves the current Lobby
func (c *Client) LeaveLobby() error {
	if c.lobby == nil {
		return errors.New("Error: Can't force start a game when not in a Lobby. Try joining a game first")
	}

	c.sendMessage(msg, "cancel")
	c.lobby = nil
	return nil
}
