package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const SERVER = "amqp://guest:guest@localhost:5672"

func main() {
	fmt.Println("Starting Peril client...")

	conn, err := amqp.Dial(SERVER)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Server connection successful!")
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		panic(err)
	}

	pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("pause.%v", username),
		routing.PauseKey,
		pubsub.TRANSIENT,
	)

	// wait for close
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Client is shutting down ...")
}
