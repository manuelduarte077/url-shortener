package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url-shortener/api/handlers"
	"url-shortener/configs"
	"url-shortener/internal/application"
	"url-shortener/internal/infrastructure/generator"
	"url-shortener/internal/infrastructure/repository"
	"url-shortener/pkg/middleware"
	"url-shortener/pkg/observability"
)

func main() {
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	urlRepo := repository.NewMemoryURLRepository(cfg.Storage.TTL)
	codeGenerator := generator.NewRandomShortCodeGenerator()
	shortenerService := application.NewShortenerService(urlRepo, codeGenerator)

	tmpl, err := template.ParseGlob("api/templates/*.html")
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	shortenerHandler := handlers.NewShortenerHandler(shortenerService, tmpl)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			shortenerHandler.ShowForm(w, r)
		} else {
			shortenerHandler.Redirect(w, r)
		}
	})
	mux.HandleFunc("/shorten", shortenerHandler.CreateShortURL)
	mux.HandleFunc("/qrcode/", shortenerHandler.GetQRCode)

	cleanupTracing, err := observability.InitTracing(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize tracing: %v", err)
	} else {
		defer cleanupTracing()
	}

	handler := middleware.RecoveryMiddleware(
		middleware.TracingMiddleware(
			middleware.LoggingMiddleware(mux),
		),
	)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if memRepo, ok := urlRepo.(*repository.MemoryURLRepository); ok {
		memRepo.Close()
	}
}
