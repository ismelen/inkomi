# Build context: back/ (copies common/)
FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache gcc musl-dev pkgconf vips-dev

WORKDIR /src
COPY common/ ./common/
COPY manga-worker/go.mod manga-worker/go.sum ./manga-worker/
WORKDIR /src/manga-worker
RUN go mod download

COPY manga-worker/ ./
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /manga-worker .


FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates tzdata vips

WORKDIR /app
COPY --from=builder /manga-worker .

EXPOSE 3003

ENTRYPOINT ["/app/manga-worker"]
