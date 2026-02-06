package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// mockScreenshotter records call args and returns configurable results.
type mockScreenshotter struct {
	url     string
	width   int
	height  int
	delayMs int

	png []byte
	err error
}

func (m *mockScreenshotter) TakeScreenshot(_ context.Context, url string, width, height, delayMs int) ([]byte, error) {
	m.url = url
	m.width = width
	m.height = height
	m.delayMs = delayMs
	return m.png, m.err
}

func TestNoURL_RendersTemplate(t *testing.T) {
	h := New(&mockScreenshotter{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Fatalf("expected text/html content-type, got %q", ct)
	}
	if body := rec.Body.String(); len(body) == 0 {
		t.Fatal("expected non-empty HTML body")
	}
}

func TestNonRootPath_404(t *testing.T) {
	h := New(&mockScreenshotter{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestPost_405(t *testing.T) {
	h := New(&mockScreenshotter{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestFTPScheme_400(t *testing.T) {
	h := New(&mockScreenshotter{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?url=ftp://example.com", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestNoScheme_PrependsHTTPS(t *testing.T) {
	m := &mockScreenshotter{png: []byte("PNG")}
	h := New(m)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?url=example.com", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if m.url != "https://example.com" {
		t.Fatalf("expected https://example.com, got %s", m.url)
	}
}

func TestDefaultWidthHeight(t *testing.T) {
	m := &mockScreenshotter{png: []byte("PNG")}
	h := New(m)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?url=https://example.com", nil)
	h.ServeHTTP(rec, req)

	if m.width != DefaultWidth {
		t.Fatalf("expected width %d, got %d", DefaultWidth, m.width)
	}
	if m.height != DefaultHeight {
		t.Fatalf("expected height %d, got %d", DefaultHeight, m.height)
	}
}

func TestCustomWidthHeightDelay(t *testing.T) {
	m := &mockScreenshotter{png: []byte("PNG")}
	h := New(m)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?url=https://example.com&width=1024&height=768&delay=500", nil)
	h.ServeHTTP(rec, req)

	if m.width != 1024 {
		t.Fatalf("expected width 1024, got %d", m.width)
	}
	if m.height != 768 {
		t.Fatalf("expected height 768, got %d", m.height)
	}
	if m.delayMs != 500 {
		t.Fatalf("expected delay 500, got %d", m.delayMs)
	}
}

func TestDelayClampedToMax(t *testing.T) {
	m := &mockScreenshotter{png: []byte("PNG")}
	h := New(m)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?url=https://example.com&delay=99999", nil)
	h.ServeHTTP(rec, req)

	if m.delayMs != MaxDelay {
		t.Fatalf("expected delay clamped to %d, got %d", MaxDelay, m.delayMs)
	}
}

func TestScreenshotSuccess_PNG(t *testing.T) {
	pngData := []byte{0x89, 'P', 'N', 'G'}
	m := &mockScreenshotter{png: pngData}
	h := New(m)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?url=https://example.com", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("expected image/png, got %q", ct)
	}
	if cl := rec.Header().Get("Content-Length"); cl != strconv.Itoa(len(pngData)) {
		t.Fatalf("expected Content-Length %d, got %s", len(pngData), cl)
	}
	if rec.Body.Len() != len(pngData) {
		t.Fatalf("expected body length %d, got %d", len(pngData), rec.Body.Len())
	}
}

func TestScreenshotError_502(t *testing.T) {
	m := &mockScreenshotter{err: errors.New("browser crashed")}
	h := New(m)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?url=https://example.com", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}
}

func TestParseIntOr(t *testing.T) {
	tests := []struct {
		s        string
		fallback int
		want     int
	}{
		{"", 42, 42},
		{"100", 42, 100},
		{"0", 42, 42},
		{"-5", 42, 42},
		{"abc", 42, 42},
		{"1", 0, 1},
	}
	for _, tt := range tests {
		got := parseIntOr(tt.s, tt.fallback)
		if got != tt.want {
			t.Errorf("parseIntOr(%q, %d) = %d, want %d", tt.s, tt.fallback, got, tt.want)
		}
	}
}
