package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const SERVER = "amqp://guest:guest@localhost:5672"

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(state routing.PlayingState) pubsub.AckType {
		defer fmt.Println(">")
		gs.HandlePause(state)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Println(">")
		outcome := gs.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		case gamelogic.MoveOutcomeMakeWar:
			return pubsub.Ack
		default:
			return pubsub.NackDiscard
		}
	}
}

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

	channel, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	// pubsub.DeclareAndBind(
	// 	conn,
	// 	routing.ExchangePerilDirect,
	// 	fmt.Sprintf("pause.%v", username),
	// 	routing.PauseKey,
	// 	pubsub.TRANSIENT,
	// )

	state := gamelogic.NewGameState(username)

	pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("pause.%v", username),
		routing.PauseKey,
		pubsub.TRANSIENT,
		handlerPause(state),
	)

	pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		fmt.Sprintf("army_moves.%v", username),
		"army_moves.*",
		pubsub.TRANSIENT,
		handlerMove(state),
	)

loop:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			// []string{"americas", "europe", "africa", "asia", "antarctica", "australia"}
			// []string{"infantry", "cavalry", "artillery"}
			err := state.CommandSpawn(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "move":
			move, err := state.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilTopic,
				fmt.Sprintf("army_moves.%v", username),
				move,
			)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("move was successful!")
		case "status":
			state.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break loop
		default:
			fmt.Println("unknown command")
			continue
		}
	}
}
