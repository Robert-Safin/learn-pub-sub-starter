package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	connection, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Error: %w", err)
	}
	defer connection.Close()
	fmt.Println("Connected to RabbitMQ")
	fmt.Println("Starting Peril server...")

	ch, err := connection.Channel()
	defer ch.Close()
	if err != nil {
		log.Fatal(err)
	}

	q, err := ch.QueueDeclare(routing.GameLogSlug, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = ch.QueueBind(routing.GameLogSlug, "game_logs.*", routing.ExchangePerilTopic, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("created and bound q", q)

	gamelogic.PrintServerHelp()

	for {
		inputs := gamelogic.GetInput()

		if len(inputs) == 0 {
			continue
		}
		switch inputs[0] {
		case "pause":
			fmt.Println("pausing")
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
		case "resume":
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
		case "quit":
			fmt.Println("quiting")
			break
		default:
			fmt.Println("unknown command")
		}

	}

	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
	// fmt.Println("Shutting down")
}
