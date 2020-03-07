package main

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
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
}

// Connect Connects to the server and returns the connected WebSocket client
func Connect(server string, userid string, username string) (*Client, error) {

	dialer := &websocket.Dialer{}
	dialer.EnableCompression = false
	url := "ws://ws.generals.io/socket.io/?EIO=3&transport=websocket"

	if server == "eu" {
		url = "ws://euws.generals.io/socket.io/?EIO=3&transport=websocket"
	}
	if server == "bot" {
		url = "ws://botws.generals.io/socket.io/?EIO=3&transport=websocket"
	}

	c, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:     c,
		userID:   userid,
		username: username,
		send:     make(chan []byte, 10),
	}, nil

}

// Run Starts the WebSocket server
func (c *Client) Run(finished chan bool) error {
	go func() {
		// I guess this is for ping-pongs?
		for range time.Tick(5 * time.Second) {
			c.send <- []byte("2")
		}
	}()
	go func() {
		time.Sleep(100 * time.Millisecond)
		for data := range c.send {
			err := c.conn.WriteMessage(websocket.TextMessage, data)
			log.Println("Sending: ", string(data))
			if err != nil {
				c.Close()
			}
		}
	}()
	c.sendMessage("set_username", c.userID, c.username)
	finished <- true
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return err
		}
		log.Println("Got: ", string(message))
		dec := json.NewDecoder(bytes.NewBuffer(message))
		var msgType int
		dec.Decode(&msgType)

		if msgType == 42 {
			var raw json.RawMessage
			dec.Decode(&raw)
			eventname := ""
			data := []interface{}{&eventname}
			json.Unmarshal(raw, &data)
			// if f, ok := c.events[eventname]; ok {
			// 	f(raw)
			// }
		}
	}
}

func (c *Client) sendMessage(v ...interface{}) {
	buf, _ := json.Marshal(v)
	newbuf := []byte("42" + string(buf))
	c.send <- newbuf
}

// Close closes the WebSocket connection
func (c *Client) Close() {
	print("closing connection")
	c.conn.Close()
}

// JoinCustomGame joins a custom game with the specified ID. Doesn't return the game object
func (c *Client) JoinCustomGame(ID string) {
	log.Println("Joined custom game at http://bot.generals.io/games/", ID)
	c.sendMessage("join_private", ID, c.userID)
	// g := &Game{c: c, ID: ID}
	// g.registerEvents()
	// return g
}
