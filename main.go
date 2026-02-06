package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kubahorak/pagecap/internal/browser"
	"github.com/kubahorak/pagecap/internal/handler"
)

const (
	defaultPort     = "8080"
	shutdownTimeout = 5 * time.Second
)

func main() {
	b, err := browser.Start(browser.DefaultTimeout)
	if err != nil {
		log.Fatalf("failed to start browser: %v", err)
	}
	log.Println("browser started")

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	h := handler.New(b)

	mux := http.NewServeMux()
	mux.Handle("/", h)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 40 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	b.Stop()
	log.Println("stopped")
}
