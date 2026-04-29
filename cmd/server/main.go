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

type UserHandler struct {
	db.UserRepository
}

func login(uh UserHandler) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}{}
		err := json.NewDecoder(r.Body).Decode(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		hPass, err := auth.HashPassword(v.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		err = uh.CreateUser(ctx, v.Username, hPass, v.Role)
		if err != nil {
			if err == context.DeadlineExceeded {
				http.Error(w, "Database is slow", http.StatusGatewayTimeout)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Username string `json:"username"`
			Password string `json:"password"`
			MatchID  int32  `json:"match_id"`
		}{}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		err := json.NewDecoder(r.Body).Decode(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		user, err := uh.GetUserByUsername(ctx, v.Username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if !auth.CheckPasswordHash(v.Password, user.PasswordHash) {
			http.Error(w, "Wrong password", http.StatusUnauthorized)
			return
		}
		token, err := auth.GenerateToken(user.ID, v.MatchID, user.Role)
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
	var uh UserHandler
	dbURL := os.Getenv("DB_URL")
	uh.D, err = db.InitDB(dbURL)
	if err != nil {
		slog.Error("Error in db connection", "Error", err)
		os.Exit(1)
	}
	defer uh.D.Close()
	err = db.RunDBMigration(".", uh.D)
	if err != nil {
		slog.Error("Error in db migration", "Error", err)
		os.Exit(1)
	}
	go login(uh)
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
