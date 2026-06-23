FROM golang:1.25.8-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go .
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /inkomi-api .



FROM alpine:3.21 AS runtime

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /inkomi-api .

# VOLUME ["/app/credentials"]

EXPOSE 3000

ENTRYPOINT ["/app/inkomi-api"]
