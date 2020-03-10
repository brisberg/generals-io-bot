// Package client provides utilities for interacting with Generals.io over a WebSocket connection.
package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/brisberg/generals-io-bot/game"
	"github.com/gorilla/websocket"
)

const (
	// Socket.io packet codes
	ping      int64  = 2
	pong      int64  = 3
	msg       int64  = 42
	serverPtn string = "ws://%vws.generals.io/socket.io/?EIO=3&transport=websocket"
)

// Client is a middleman between the websocket connection and client application or bot.
type Client struct {
	// The websocket connection.
	conn *websocket.Conn
	// Session ID of current connection
	sid string

	// Current User object
	user *User

	// Current Lobby
	lobby *Lobby

	// Current Game
	Game *game.Game

	// Buffered channel of outbound messages.
	send chan []byte

	// Buffered channel of pong responses
	pong chan bool

	// Channel indicating the connection is closed and we should clean up
	closed chan bool

	// Callback for when the connection is closed
	OnClose func()
}

type connConfig struct {
	SID          string `json:"sid"` // session ID
	PingInterval int    `json:"pingInterval"`
	PingTimeout  int    `json:"pingTimeout"`
}

// Connect Connects to the server and returns the connected WebSocket client
//
// Server param should be one of "" = US, "es" = Europe, "bot" = Bot (SF) server
func Connect(server string) (*Client, error) {

	dialer := &websocket.Dialer{}
	dialer.EnableCompression = false
	url := fmt.Sprintf(serverPtn, server)

	// Dial the server
	c, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	// Read connection config from server
	// Expect msg type to be `0` (open)
	_, configMsg, err := c.ReadMessage()
	if err != nil {
		return nil, err
	}
	var msgType int
	config := connConfig{}

	log.Println("Got: ", string(configMsg))
	decodeSocketIoMessage(configMsg, msgType, &config)

	// Read success message from server
	// Expect msg type to be `40`
	_, message, err := c.ReadMessage()
	if err != nil {
		return nil, err
	}

	if string(message) != "40" {
		return nil, errors.New(fmt.Sprint("Error: Expected '40' success type: got ", msgType))
	}
	log.Println("Connection Established.")

	user := &User{
		usernamec: make(chan string, 1),
		usererrc:  make(chan string, 1),
	}

	client := &Client{
		conn:   c,
		user:   user,
		sid:    config.SID,
		send:   make(chan []byte, 10),
		pong:   make(chan bool, 1),
		closed: make(chan bool),
	}

	go client.schedulePingPong(&config)

	return client, nil
}

// Set up repeated ping requests to the server
// Respects the pingInterval and pingTimeout provided by the server when opening the connection
// If a pong ("3") is not recieved before the timeout, the server is assumed nonresponsive and
// the connection is closed
func (c *Client) schedulePingPong(config *connConfig) {
	ticker := time.NewTicker(time.Duration(config.PingInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send a ping
			c.send <- []byte(strconv.FormatInt(ping, 10) + " ping")
			timeout := time.After(time.Duration(config.PingTimeout) * time.Millisecond)
			select {
			case <-c.pong:
				// Pong satisfied, do nothing
			case <-timeout:
				c.Close("Error Pong Timeout. Connection Lost.")
				return
			}
		case <-c.closed:
			return
		}
	}
}

// Decodes the packet type and the data from a socket.io message
// TODO: Refactor this to be more flexible to multiplex on the packettype
func decodeSocketIoMessage(msg []byte, msgType int, data interface{}) error {
	dec := json.NewDecoder(bytes.NewBuffer(msg))
	if err := dec.Decode(&msgType); err != nil {
		return err
	}
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return err
	}
	// log.Println(raw)
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}
	// log.Println(data)

	return nil
}

// Run Starts the WebSocket server
func (c *Client) Run() error {
	// Launch goroutine to process outbound requests
	go func() {
		time.Sleep(100 * time.Millisecond)
		for data := range c.send {
			err := c.conn.WriteMessage(websocket.TextMessage, data)
			log.Println("Sending: ", string(data))
			if err != nil {
				c.Close(fmt.Sprint("Error Sending Request: ", err))
			}
		}
	}()

	// Loop and process inbound responses
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return err
		}
		log.Println("Got: ", string(message))
		dec := json.NewDecoder(bytes.NewBuffer(message))
		var msgType int64
		dec.Decode(&msgType)

		// Dispatch the message to various listeners
		if msgType == msg {
			var raw json.RawMessage
			dec.Decode(&raw)
			eventname := ""
			data := []interface{}{&eventname}
			json.Unmarshal(raw, &data)
			// log.Println(eventname)
			// log.Println(data)
			// if f, ok := c.events[eventname]; ok {
			// 	f(raw)
			// }
			if eventname == "pre_game_start" {
				c.Game = &game.Game{ID: "foobar"}
			} else if eventname == "game_start" || eventname == "game_update" {
				c.Game.Dispatch(eventname, raw)
			} else if eventname == "game_over" {
				c.sendMessage(msg, "leave_game")
				c.Game = nil
				c.Close("Game concluded.")
			} else if eventname == "error_set_username" {
				// TODO: split this into the user package
				// Unwrap the error_set_username event and pass back to user
				data := []string{}
				json.Unmarshal(raw, &data)
				c.user.usererrc <- data[1]
			}
		} else if msgType == pong {
			c.pong <- true
		} else if msgType == 430 {
			// TODO split this into the user package
			var raw json.RawMessage
			dec.Decode(&raw)
			var name []string
			json.Unmarshal(raw, &name)
			c.user.usernamec <- name[0]
		}
	}
}

// Sends a message to the GameServer over the WebSocket
func (c *Client) sendMessage(code int64, v ...interface{}) {
	buf, _ := json.Marshal(v)
	newbuf := []byte(strconv.FormatInt(code, 10) + string(buf))
	c.send <- newbuf
}

// Close closes the WebSocket connection
func (c *Client) Close(msg string) {
	log.Println(msg)
	log.Println("Closing client connection...")
	c.conn.Close()
	close(c.closed)
	close(c.pong)

	if c.OnClose != nil {
		c.OnClose()
	}
}

// Attack sends an attack request to the server
func (c *Client) Attack(from, to int, is50 bool) {
	c.sendMessage(msg, "attack", from, to, is50, c.Game.NextAttackIndex())
}
