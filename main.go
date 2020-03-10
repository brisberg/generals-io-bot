package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/brisberg/generals-io-bot/client"
)

func main() {
	fmt.Printf("Starting Generals AI Program:\n")

	client, err := client.Connect("bot")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close("Main Program Terminated")

	client.OnClose = func() {
		log.Println("Close Callback. Exiting.")
		os.Exit(1)
	}

	go client.Run()

	if err := client.RegisterBot("mybot-batz", "[Bot]Keidence-45"); err != nil {
		log.Fatalln(err)
	}
	time.Sleep(2000 * time.Millisecond)

	client.JoinCustomGame("botbotbot")
	time.Sleep(3000 * time.Millisecond)

	client.SetForceStart(true)
	time.Sleep(3000 * time.Millisecond)

	for {
		time.Sleep(5000 * time.Millisecond)
	}

	// client.LeaveLobby()
	// time.Sleep(10000 * time.Millisecond)
}
