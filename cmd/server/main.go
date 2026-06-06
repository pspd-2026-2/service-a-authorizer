package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"service-a/config"
	"service-a/internal/authorizer"
	"syscall"
	"time"

	"service-a/internal/pb"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	svc := authorizer.New()
	httpHandler := authorizer.NewHandler(svc)
	grpcHandler := authorizer.NewGRPCHandler(svc)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      loggingMiddleware(setupHTTPRoutes(httpHandler)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[BOOT] %s (HTTP REST) listening on http://localhost:%s", cfg.ServiceName, cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERROR] HTTP server error: %v", err)
		}
	}()

	grpcServer := grpc.NewServer()
	registerAuthorizationService(grpcServer, grpcHandler)

	go func() {
		listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("[ERROR] gRPC listener error: %v", err)
		}

		log.Printf("[BOOT] %s (gRPC) listening on grpc://localhost:%s", cfg.ServiceName, cfg.GRPCPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("[ERROR] gRPC server error: %v", err)
		}
	}()

	<-quit
	log.Println("[SHUTDOWN] signal received, shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("[SHUTDOWN] HTTP forced shutdown: %v", err)
	}

	grpcServer.GracefulStop()

	log.Println("[SHUTDOWN] servers stopped cleanly")
}

func setupHTTPRoutes(handler *authorizer.Handler) http.Handler {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return mux
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[HTTP] %s %s — %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func registerAuthorizationService(s *grpc.Server, impl pb.AuthorizationServiceServer) {
	pb.RegisterAuthorizationServiceServer(s, impl)
}
