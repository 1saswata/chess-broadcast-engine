package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/1saswata/chess-broadcast-engine/internal/cache"
	"github.com/1saswata/chess-broadcast-engine/internal/pb"
	"github.com/1saswata/chess-broadcast-engine/internal/telemetry"
	"github.com/1saswata/chess-broadcast-engine/internal/websocket"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	tp, err := telemetry.InitTracer("chess-broadcast-server")
	defer tp.Shutdown(context.Background())
	if err != nil {
		slog.Error("Error creating tracer", "Error", err)
		os.Exit(1)
	}
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
	rc, err := cache.NewRedisCache("localhost:6379")
	if err != nil {
		slog.Error("Error connecting to cache", "Error", err)
		os.Exit(1)
	}
	defer rc.Close()
	wsHandler := websocket.NewWsHandler(hub, rc)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", wsHandler.ServeHttp)
	newServer := http.Server{Addr: ":8081", Handler: mux}
	go func() {
		if err := newServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP Server error:", "Error", err)
			os.Exit(1)
		}
	}()
	go func() {
		for d := range msgs {
			func() {
				carrier := telemetry.AMQPCarrier{Table: d.Headers}
				ctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)
				ctx, span := otel.Tracer("broadcaster-node").Start(ctx, "BroadcastMove")
				defer span.End()
				var move pb.Move
				err = proto.Unmarshal(d.Body, &move)
				if err != nil {
					slog.Error("Error serializing move", "Error", err)
					return
				}
				slog.Info("Move", "Staring square", move.StartingSquare,
					"Destination square", move.DestinationSquare)
				jsonByte, err := protojson.Marshal(&move)
				if err != nil {
					slog.Error("Error converting to json bytes ", "Error", err)
				} else {
					hub.Broadcast(&websocket.BroadcastMessage{
						MatchID: move.MatchId,
						Payload: jsonByte,
					})
				}
			}()
		}
	}()
	wait := make(chan os.Signal, 1)
	signal.Notify(wait, syscall.SIGINT, syscall.SIGTERM)
	<-wait
	slog.Info("Closing down the server")
}
