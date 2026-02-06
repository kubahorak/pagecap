# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY internal/ ./internal/
RUN go build -o /pagecap .

# Build the Playwright CLI for use in the runtime stage
RUN go build -o /playwright-cli github.com/playwright-community/playwright-go/cmd/playwright

# Stage 2: Runtime
FROM ubuntu:24.04

# ca-certificates is needed for the playwright CLI to download over HTTPS
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install WebKit and its system deps via Playwright's own installer
COPY --from=builder /playwright-cli /tmp/playwright-cli
RUN /tmp/playwright-cli install --with-deps webkit \
    && rm /tmp/playwright-cli \
    && apt-get update && apt-get install -y --no-install-recommends \
       fonts-liberation fonts-noto-color-emoji \
    && rm -rf /var/lib/apt/lists/*

RUN useradd -m pagecap

COPY --from=builder /pagecap /usr/local/bin/pagecap
RUN mv /root/.cache /home/pagecap/.cache \
    && chown -R pagecap:pagecap /home/pagecap/.cache

USER pagecap
EXPOSE 8080
CMD ["pagecap"]
