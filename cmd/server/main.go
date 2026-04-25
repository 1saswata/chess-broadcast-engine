package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/1saswata/chess-broadcast-engine/internal/auth"
	"github.com/1saswata/chess-broadcast-engine/internal/broker"
	"github.com/1saswata/chess-broadcast-engine/internal/cache"
	"github.com/1saswata/chess-broadcast-engine/internal/db"
	"github.com/1saswata/chess-broadcast-engine/internal/pb"
	"github.com/1saswata/chess-broadcast-engine/internal/server"
	"github.com/1saswata/chess-broadcast-engine/internal/telemetry"
	"google.golang.org/grpc"
)

func login() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Role     string `json:"role"`
			Match_id int32  `json:"match_id"`
		}{}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		err := json.NewDecoder(r.Body).Decode(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		token, err := auth.GenerateToken(v.Match_id, v.Role)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "%s", token)
	})
	newServer := http.Server{Addr: ":8080", Handler: mux}
	slog.Error("Error on login", "Error", newServer.ListenAndServe())
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	tp, err := telemetry.InitTracer("chess-ingest-server")
	if err != nil {
		slog.Error("Error creating tracer", "Error", err)
		os.Exit(1)
	}
	dbURL := os.Getenv("DB_URL")
	d, err := db.InitDB(dbURL)
	if err != nil {
		slog.Error("Error in db connection", "Error", err)
		os.Exit(1)
	}
	defer d.Close()
	err = db.RunDBMigration(".", d)
	if err != nil {
		slog.Error("Error in db migration", "Error", err)
		os.Exit(1)
	}
	go login()
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
	s := grpc.NewServer(grpc.UnaryInterceptor(server.AuthInterceptor))
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
