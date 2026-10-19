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

	"go-task-api/internal/database"
	"go-task-api/internal/task"
	"go-task-api/internal/user"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	db, err := database.Open(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))

	})

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	taskRepo := task.NewRepository(db)
	taskService := task.NewService(taskRepo)
	taskHandler := task.NewHandler(taskService)

	mux.HandleFunc("POST /register", userHandler.Register)
	mux.HandleFunc("POST /login", userHandler.Login)

	mux.Handle("POST /tasks",
		userHandler.AuthMiddleware(http.HandlerFunc(taskHandler.Create)))
	mux.Handle("GET /tasks",
		userHandler.AuthMiddleware(http.HandlerFunc(taskHandler.List)))
	mux.Handle("GET /tasks/{id}",
		userHandler.AuthMiddleware(http.HandlerFunc(taskHandler.Get)))
	mux.Handle("PUT /tasks/{id}",
		userHandler.AuthMiddleware(http.HandlerFunc(taskHandler.Update)))
	mux.Handle("DELETE /tasks/{id}",
		userHandler.AuthMiddleware(http.HandlerFunc(taskHandler.Delete)))

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("server started", "port", port)
		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}
