package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"service-a/config"
	"service-a/internal/authorizer"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()
 
	log.Printf("[BOOT] starting %s on port %s", cfg.ServiceName, cfg.HTTPPort)
 
	// Wiring: repositório → use case → handler
	svc := authorizer.New()
	handler := authorizer.NewHandler(svc)
 
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
 
	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
 
	// Graceful shutdown: aguarda SIGINT / SIGTERM (importante para Kubernetes)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
 
	go func() {
		log.Printf("[BOOT] %s listening on http://localhost:%s", cfg.ServiceName, cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERROR] server error: %v", err)
		}
	}()
 
	<-quit
	log.Println("[SHUTDOWN] signal received, shutting down gracefully...")
 
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
 
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("[SHUTDOWN] forced shutdown: %v", err)
	}
 
	log.Println("[SHUTDOWN] server stopped cleanly")
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[HTTP] %s %s — %s", r.Method, r.URL.Path, time.Since(start))
	})
}