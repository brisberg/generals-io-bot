package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	fmt.Printf("Starting Generals AI Program:\n")

	client, err := Connect("bot", "mybot-batz", "[Bot]Keidence-45")
	if err != nil {
		log.Fatal(err)
	}

	client.OnClose = func() {
		log.Println("Close Callback. Exiting.")
		os.Exit(1)
	}

	finished := make(chan bool)
	go client.Run(finished)

	// go func() {
	// 	_, ok := <-client.closed
	// 	if !ok {
	// 		os.Exit(1)
	// 	}
	// }()

	<-finished
	client.JoinCustomGame("botbotbot")

	time.Sleep(3000)
	client.sendMessage("set_force_start", "botbotbot", true)

	for {
		time.Sleep(1000)
	}
}
