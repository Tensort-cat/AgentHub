FROM golang:1.26.7 AS builder

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/agenthub ./cmd

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/agenthub /app/agenthub
COPY --from=builder /src/configs /app/configs
COPY --from=builder /src/static /app/static
COPY --from=builder /src/logs /app/logs

EXPOSE 8000

CMD ["/app/agenthub"]
