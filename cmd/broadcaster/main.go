package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/1saswata/chess-broadcast-engine/internal/pb"
	"github.com/1saswata/chess-broadcast-engine/internal/websocket"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/encoding/protojson"
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
	if err != nil {
		slog.Error("Error declaring queue", "Error", err)
		os.Exit(1)
	}
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
	hub := websocket.NewHub()
	go hub.Run()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws",
		func(w http.ResponseWriter, r *http.Request) { websocket.ServeWs(hub, w, r) })
	newServer := http.Server{Addr: ":8081", Handler: mux}
	go func() {
		if err := newServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP Server error:", "Error", err)
			os.Exit(1)
		}
	}()
	go func() {
		for d := range msgs {
			var move pb.Move
			err = proto.Unmarshal(d.Body, &move)
			if err != nil {
				slog.Error("Error serializing move", "Error", err)
			}
			slog.Info("Move", "Staring square", move.StartingSquare,
				"Destination square", move.DestinationSquare)
			jsonByte, err := protojson.Marshal(&move)
			if err != nil {
				slog.Error("Error converting to json bytes ", "Error", err)
			} else {
				hub.Broadcast(jsonByte)
			}
		}
	}()
	wait := make(chan os.Signal, 1)
	signal.Notify(wait, syscall.SIGINT, syscall.SIGTERM)
	<-wait
	slog.Info("Closing down the server")
}
