package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Socket.io packet codes
	ping      int64  = 2
	pong      int64  = 3
	msg       int64  = 42
	serverPtn string = "ws://%vws.generals.io/socket.io/?EIO=3&transport=websocket"
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// The websocket connection.
	conn *websocket.Conn

	// The secret UserId for the bot account
	userID string
	// Public Username for the bot
	username string

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

// Create dialer, and determine the URL
// Dial the server
// Read one message to get back the connection config
// 	Read out the SID, PingInterval, and PingTimeout
// 	Start a ticker with the Ping Interval
// 		Each time the ticker goes off, send a `2` message
// 		Wait up to PingTimeout
// 			If we haven't returned a `3`, assume server is down. Close connection
// 			If we did, wait for next timer tick
// 	When the client shuts down, stop the timer
// Read a second message to confirm it is '40'
// Return fully connected Client

// Connect Connects to the server and returns the connected WebSocket client
//
// Server param should be one of "" = US, "es" = Europe, "bot" = Bot (SF) server
func Connect(server string, userid string, username string) (*Client, error) {

	dialer := &websocket.Dialer{}
	dialer.EnableCompression = false
	url := fmt.Sprintf(serverPtn, server)

	// Dial the server
	c, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	// Read connection config from server
	// Expect msg type to be `0`
	_, message, err := c.ReadMessage()
	if err != nil {
		return nil, err
	}
	var msgType int
	config := connConfig{}

	log.Println("Got: ", string(message))
	decodeSocketIoMessage(message, msgType, &config)

	client := &Client{
		conn:     c,
		userID:   userid,
		username: username,
		send:     make(chan []byte, 10),
		pong:     make(chan bool, 1),
		closed:   make(chan bool),
	}

	client.schedulePingPong(&config)

	return client, nil
}

func (c *Client) schedulePingPong(config *connConfig) {
	ticker := time.NewTicker(time.Duration(config.PingInterval) * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				// Send a ping
				c.send <- []byte(strconv.FormatInt(ping, 10) + " ping")
				log.Println("Sending Ping")
				timeout := time.After(time.Duration(config.PingTimeout) * time.Millisecond)
				select {
				case <-c.pong:
					// Pong satisfied, do nothing
					log.Println("Pong back, recieved")
				case <-timeout:
					log.Println("Timeout expired")
					c.Close("Error Pong Timeout. Connection Lost.")
					return
				}
			case <-c.closed:
				ticker.Stop()
				return
			}
		}
	}()
}

func decodeSocketIoMessage(msg []byte, msgType int, data interface{}) error {
	dec := json.NewDecoder(bytes.NewBuffer(msg))
	if err := dec.Decode(&msgType); err != nil {
		return err
	}
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}

	return nil
}

// Run Starts the WebSocket server
func (c *Client) Run(finished chan bool) error {
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
	finished <- true
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return err
		}
		log.Println("Got: ", string(message))
		dec := json.NewDecoder(bytes.NewBuffer(message))
		var msgType int64
		dec.Decode(&msgType)

		if msgType == msg {
			var raw json.RawMessage
			dec.Decode(&raw)
			eventname := ""
			data := []interface{}{&eventname}
			json.Unmarshal(raw, &data)
			// if f, ok := c.events[eventname]; ok {
			// 	f(raw)
			// }
			if eventname == "game_over" {
				c.sendMessage("leave_game")
				c.Close("Game concluded.")
			}
		} else if msgType == pong {
			c.pong <- true
		}
	}
}

// Sends a message to the GameServer over the WebSocket
func (c *Client) sendMessage(v ...interface{}) {
	buf, _ := json.Marshal(v)
	newbuf := []byte(strconv.FormatInt(msg, 10) + string(buf))
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

// JoinCustomGame joins a custom game with the specified ID. Doesn't return the game object
func (c *Client) JoinCustomGame(ID string) {
	log.Println("Joined custom game at http://bot.generals.io/games/", ID)
	c.sendMessage("join_private", ID, c.userID)
	// g := &Game{c: c, ID: ID}
	// g.registerEvents()
	// return g
}
