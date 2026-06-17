FROM golang:alpine3.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /ermc-api main.go


FROM alpine:edge
WORKDIR /root/

COPY --from=builder /ermc-api .

EXPOSE 8080

ENTRYPOINT ["./ermc-api", "serve", "--port", "8080"]