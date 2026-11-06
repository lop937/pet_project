package server

import (
	"log"
	"net/http"
	"time"
)

// Config — конфигурация сервера.
type Config struct {
	Url string
}

// Serve запускает HTTP-сервер на cfg.Url.
// tn — время старта, используется для логирования.
func Serve(cfg Config, tn time.Time, handler http.Handler) {
	log.Printf("server: starting on %s (since %s)", cfg.Url, time.Since(tn))
	if err := http.ListenAndServe(cfg.Url, handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}
