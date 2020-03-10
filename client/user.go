// Package client provides utilities for interacting with Generals.io over a WebSocket connection.
//
// User adds types and utility functions for managing a User presense on Generals.io.
package client

import (
	"fmt"
	"time"
)

// User holds information about a user or botuser on Generals.io
type User struct {
	// The secret UserId for the bot account
	userID string
	// Public Username for the bot
	username string
	// rank int
	// stars int

	// Buffered Channel for the results of a get_username request
	usernamec chan string
	// Buffered Channel for the results of a set_username request
	usererrc chan string
}

// RegisterBot verifies that the given UserID is associated with the resired username.
// If it is not, attmpts to update the user name to the given one.
//
// This will fail if the username doesn't start with `[BOT]`, the username is already
// taken by another user, or the given userID is already associated with a different
// username.
func (c *Client) RegisterBot(userID string, username string) error {
	if userID == "" || username == "" {
		return fmt.Errorf("Error: Must specify both a userID and a username to register a bot")
	}

	// Fetch the current username
	curName, err := c.getUsername(userID)
	if err != nil {
		return err
	}

	// If username is different than desired name, update it
	if curName != username {
		if err := c.setUsername(userID, username); err != nil {
			return err
		}
	}

	c.user.userID = userID
	c.user.username = username
	return nil
}

func (c *Client) getUsername(userID string) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("Error: Could not fetch Username without a UserID")
	}

	c.sendMessage(420, "get_username", userID)

	// Block on waiting for the message dispatch
	select {
	case n := <-c.user.usernamec:
		return n, nil
	case <-time.After(20 * time.Second):
		return "", fmt.Errorf("Error: Could not fetch Username. Server never replied with it")
	}
}

type getUserNameResp []string

func (c *Client) handleGetUsername() {}

type setUserNameResp string

func (c *Client) setUsername(userID string, username string) error {
	// Send the Username change
	c.sendMessage(msg, "set_username", c.user.userID, c.user.username)

	// Block on waiting for the message dispatch
	select {
	case e := <-c.user.usererrc:
		if e == "" {
			return nil
		}
		return fmt.Errorf("Error: Could not register bot under username %v: %v", username, e)
	case <-time.After(20 * time.Second):
		return fmt.Errorf("Error: Could not register bot under new username. Server timeout")
	}
}
