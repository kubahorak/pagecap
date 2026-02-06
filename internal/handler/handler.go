package handler

import (
	"context"
	"embed"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	DefaultWidth  = 640
	DefaultHeight = 480
	MaxDelay      = 10000
)

//go:embed templates/index.html
var templateFS embed.FS

var uiTemplate = template.Must(template.ParseFS(templateFS, "templates/index.html"))

// Screenshotter takes a screenshot of a URL.
type Screenshotter interface {
	TakeScreenshot(ctx context.Context, url string, width, height, delayMs int) ([]byte, error)
}

// Handler serves the PageCap HTTP API.
type Handler struct {
	screenshotter Screenshotter
}

// New creates a Handler that delegates screenshots to s.
func New(s Screenshotter) *Handler {
	return &Handler{screenshotter: s}
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawURL := r.URL.Query().Get("url")
	if rawURL == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := uiTemplate.Execute(w, nil); err != nil {
			log.Printf("template error: %v", err)
		}
		return
	}

	// Auto-add scheme if missing.
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		http.Error(w, "only http and https URLs are allowed", http.StatusBadRequest)
		return
	}

	width := parseIntOr(r.URL.Query().Get("width"), DefaultWidth)
	height := parseIntOr(r.URL.Query().Get("height"), DefaultHeight)
	delayMs := min(parseIntOr(r.URL.Query().Get("delay"), 0), MaxDelay)

	log.Printf("screenshot url=%s width=%d height=%d delay=%d", rawURL, width, height, delayMs)

	png, err := h.screenshotter.TakeScreenshot(r.Context(), rawURL, width, height, delayMs)
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(png)))
	_, _ = w.Write(png)
}

func parseIntOr(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}
