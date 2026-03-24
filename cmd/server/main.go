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
	"google.golang.org/grpc"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	rp, err := broker.NewRabbitMQPublisher("amqp://guest:guest@localhost:5672/")
	if err != nil {
		slog.Error("Error creating RabbitMQPublisher", "Error", err)
		os.Exit(1)
	}
	mc, err := cache.NewRedisCache("localhost:6379")
	ingestServer := server.NewIngestServer(rp, mc)
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
		err = mc.Close()
		if err != nil {
			slog.Error("Error closing redis connection", "Error", err)
		}
		err = rp.Close()
		if err != nil {
			slog.Error("Error closing rabbitMQ connection", "Error", err)
		}
		s.GracefulStop()
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
