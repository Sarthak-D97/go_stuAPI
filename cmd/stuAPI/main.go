package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sarthak-D97/go_stuAPI/internal/config"
	"github.com/Sarthak-D97/go_stuAPI/internal/http/handlers/student"
	"github.com/Sarthak-D97/go_stuAPI/internal/storage/sqlite"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Load configuration
	cfg := config.MustLoad()

	// 2. Database setup (SQLite)
	storage, err := sqlite.New(cfg)
	if err != nil {
		log.Fatal(err)
		return
	}
	slog.Info("SQLite connected successfully", slog.String("env", cfg.Env))

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	// 3. Redis client setup
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	ping, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		slog.Error("Redis connection failed", "error", err)
	} else {
		slog.Info("Redis connected successfully", "ping", ping)
	}

	// 4. Router setup
	router := http.NewServeMux()
	router.HandleFunc("POST /api/students", student.New(storage, rdb))
	router.HandleFunc("GET /api/students/{id}", student.GetById(storage, rdb))
	router.HandleFunc("PUT /api/students/{id}", student.UpdateStudent(storage, rdb))
	router.HandleFunc("GET /api/students/", student.GetList(storage, rdb))
	router.HandleFunc("DELETE /api/students/{id}", student.DeleteStudent(storage, rdb))
	// 5. Server setup
	server := http.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: router,
	}

	slog.Info("Starting server", "address", cfg.HTTPServer.Addr)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %s", err.Error())
		}
	}()

	<-done
	slog.Info("Shutting down the server")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := rdb.Close(); err != nil {
		slog.Error("Failed to close Redis", "error", err)
	}
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shut down server", "error", err)
	}

	slog.Info("Server exited properly")
}
