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
)

func main() {
	// fmt.Println("Hello, Student API!")
	//load configuration
	cfg := config.MustLoad()

	//database setup
	storage, err := sqlite.New(cfg)
	if err != nil {
		log.Fatal(err)
		return
	}
	slog.Info("Database connected successfully", slog.String("env", cfg.Env), slog.String("version", "1.0.0"))

	//router setup
	router := http.NewServeMux()
	router.HandleFunc("POST /api/students", student.New(storage))
	router.HandleFunc("GET /api/students/{id}", student.GetById(storage))
	router.HandleFunc("PUT /api/students/{id}", student.UpdateStudent(storage))
	router.HandleFunc("GET /api/students/", student.GetList(storage))
	router.HandleFunc("DELETE /api/students/{id}", student.DeleteStudent(storage))

	//setup server
	server := http.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: router,
	}
	slog.Info("Starting server", "address", cfg.HTTPServer.Addr)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
	}()

	<-done
	slog.Info("Shutting down the server")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shut down server", "error", err)
	}
	slog.Info("Server exited properly")
}
