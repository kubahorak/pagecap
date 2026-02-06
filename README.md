# PageCap

Website screenshot service for Docker.
Uses [Playwright](https://playwright.dev/) with headless Chromium to capture screenshots via an HTTP API.

Inspired by [Manet](https://github.com/vbauer/manet).

## Quick Start

### Docker

```bash
docker build -t pagecap .
docker run -p 8080:8080 pagecap
```

### Local

```bash
go run github.com/playwright-community/playwright-go/cmd/playwright install --with-deps chromium
go run .
```

Open http://localhost:8080/ for the web UI.

## API

```
GET /?url=https://example.com
GET /?url=https://example.com&width=1280&height=960
GET /?url=https://example.com&delay=2000
```

| Parameter | Default | Description                   |
|-----------|---------|-------------------------------|
| `url`     | —       | Target URL (required)         |
| `width`   | 640     | Viewport width in pixels      |
| `height`  | 480     | Viewport height in pixels     |
| `delay`   | 0       | Wait time in ms after load    |

Returns a PNG image. URLs without a scheme get `https://` prepended automatically.

## Configuration

| Env var | Default | Description  |
|---------|---------|--------------|
| `PORT`  | 8080    | Listen port  |
