package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const SERVER = "amqp://guest:guest@localhost:5672"

func handlerLogs() func(routing.GameLog) pubsub.AckType {
	return func(gamelog routing.GameLog) pubsub.AckType {
		defer fmt.Println(">")
		fmt.Println(gamelog)
		gamelogic.WriteLog(gamelog)
		return pubsub.Ack
	}
}

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

	pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		fmt.Sprintf("%v.*", routing.GameLogSlug),
		pubsub.DURABLE,
		handlerLogs(),
	)

	gamelogic.PrintServerHelp()

loop:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "pause":
			fmt.Println("Sending pause message ...")
			pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				})
		case "resume":
			fmt.Println("Sending resume message ...")
			pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				})
		case "quit":
			fmt.Println("Server is shutting down ...")
			break loop
		default:
			fmt.Print("unknown command")
			continue
		}

	}

}
