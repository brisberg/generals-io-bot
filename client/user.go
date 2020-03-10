// Package client provides utilities for interacting with Generals.io over a WebSocket connection.
//
// User adds types and utility functions for managing a User presense on Generals.io.
package client

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/gorilla/websocket"
)

// User holds information about a user or botuser on Generals.io
type User struct {
	userID   string
	username string
	// rank int
	// stars int
}

type setUserNameResp string

func (c *Client) registerBotUsername() error {
	// Send the Username change
	c.sendMessage("set_username", c.userID, c.username)
	buf, _ := json.Marshal([]string{"set_username", c.userID, c.username})
	newbuf := []byte(strconv.FormatInt(msg, 10) + string(buf))
	if err := c.conn.WriteMessage(websocket.TextMessage, newbuf); err != nil {
		return err
	}

	// Check the error username response
	// Empty means success, "User name in use" assume that that means we own it
	_, usernameMsg, err := c.conn.ReadMessage()
	if err != nil {
		return err
	}
	var msgType int
	var nameError string

	log.Println("Got: ", string(usernameMsg))
	decodeSocketIoMessage(usernameMsg, msgType, &nameError)

	log.Println(nameError)

	if nameError == "" || nameError == "This username is already taken." {
		return nil
	}

	return fmt.Errorf("Error: Could not register bot under username %v (%v)", c.username, nameError)
}
