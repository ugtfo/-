package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "log-parser/internal/config"
    "log-parser/internal/handler"
    "log-parser/internal/repository"
    "log-parser/internal/service"
)

func main() {
    cfg := config.Load()
    
    logger := log.New(os.Stdout, "[LOG-PARSER] ", log.Ldate|log.Ltime|log.LUTC|log.Lmsgprefix)
    
    repo, err := repository.NewPostgresRepo(cfg.DatabaseURL, logger)
    if err != nil {
        logger.Fatalf("Failed to connect to database: %v", err)
    }
    defer repo.Close()

    if err := repo.RunMigrations(); err != nil {
        logger.Fatalf("Failed to run migrations: %v", err)
    }

    svc := service.NewService(repo, logger)
    h := handler.NewHandler(svc, logger)

    mux := http.NewServeMux()
    mux.HandleFunc("/api/v1/parse/", h.HandleParse)
    mux.HandleFunc("/api/v1/topology/", h.HandleTopology)
    mux.HandleFunc("/api/v1/node/", h.HandleNode)
    mux.HandleFunc("/api/v1/port/", h.HandlePort)
    mux.HandleFunc("/api/v1/log/", h.HandleLog)

    server := &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      mux,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    go func() {
        logger.Printf("Starting server on port %s", cfg.Port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatalf("Server failed: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Println("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        logger.Fatalf("Server forced to shutdown: %v", err)
    }

    logger.Println("Server stopped")
}
