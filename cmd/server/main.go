package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/1saswata/chess-broadcast-engine/internal/broker"
	"github.com/1saswata/chess-broadcast-engine/internal/cache"
	"github.com/1saswata/chess-broadcast-engine/internal/pb"
	"github.com/1saswata/chess-broadcast-engine/internal/server"
	"github.com/1saswata/chess-broadcast-engine/internal/telemetry"
	"google.golang.org/grpc"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	tp, err := telemetry.InitTracer("chess-ingest-server")
	if err != nil {
		slog.Error("Error creating tracer", "Error", err)
		os.Exit(1)
	}
	amqpURL := os.Getenv("AMQP_SERVER_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}
	rp, err := broker.NewRabbitMQPublisher(amqpURL)
	if err != nil {
		slog.Error("Error creating RabbitMQPublisher", "Error", err)
		os.Exit(1)
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	rc, err := cache.NewRedisCache(redisURL)
	if err != nil {
		slog.Error("Error connecting to cache", "Error", err)
		os.Exit(1)
	}
	ingestServer := server.NewIngestServer(rp, rc)
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("Error creating Listener", "Error", err)
		os.Exit(1)
	}
	s := grpc.NewServer()
	pb.RegisterChessIngestServiceServer(s, ingestServer)
	slog.Info("Starting the server...")
	go func() {
		if err := s.Serve(lis); err != nil {
			slog.Error("Error serving connection", "Error", err)
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	slog.Info("Server is shutting down... ")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := make(chan int, 1)
	go func() {
		err := rc.Close()
		if err != nil {
			slog.Error("Error closing redis connection", "Error", err)
		}
		err = rp.Close()
		if err != nil {
			slog.Error("Error closing rabbitMQ connection", "Error", err)
		}
		s.GracefulStop()
		err = tp.Shutdown(ctx)
		if err != nil {
			slog.Error("Error closing tracer", "Error", err)
		}
		close(done)
	}()
	select {
	case <-done:
		slog.Info("Server closed Gracefully")
	case <-ctx.Done():
		slog.Info("Graceful shutdown timeout, force closing the server...")
		s.Stop()
	}
}
