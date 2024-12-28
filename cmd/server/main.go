package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const SERVER = "amqp://guest:guest@localhost:5672"

func main() {
	fmt.Println("Starting Peril server...")

	conn, err := amqp.Dial(SERVER)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Server connection successful!")

	channel, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})

	// wait for close
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Server is shutting down ...")
}
