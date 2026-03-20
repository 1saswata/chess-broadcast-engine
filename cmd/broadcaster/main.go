package main

import (
	"log/slog"
	"os"

	"github.com/1saswata/chess-broadcast-engine/internal/pb"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		slog.Error("Error connecting to rabbitmq", "Error", err)
		os.Exit(1)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		slog.Error("Unable to create channel", "Error", err)
		os.Exit(1)
	}
	defer ch.Close()
	err = ch.ExchangeDeclare(
		"chess_broadcast",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		slog.Error("Error declaring exchange", "Error", err)
		os.Exit(1)
	}
	q, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	err = ch.QueueBind(
		q.Name,
		"",
		"chess_broadcast",
		false,
		nil,
	)
	if err != nil {
		slog.Error("Error binding queue", "Error", err)
		os.Exit(1)
	}
	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		slog.Error("Error registering consumer", "Error", err)
		os.Exit(1)
	}
	move := pb.Move{}
	for d := range msgs {
		err = proto.Unmarshal(d.Body, &move)
		if err != nil {
			slog.Error("Error serializing move", "Error", err)
		}
		slog.Info("Move", "Staring square", move.StartingSquare,
			"Destination square", move.DestinationSquare)
	}
}
