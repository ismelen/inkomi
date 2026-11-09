# Build context: back/monolith/. cgo + libvips (manga module) can't cross-compile, so every stage
# runs on the target platform (emulated with QEMU in the pipeline).
FROM golang:1.26.2-alpine AS deps
RUN apk add --no-cache gcc musl-dev pkgconf vips-dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./

# Unit tests with libvips: docker build --target test
FROM deps AS test
RUN go vet ./... && go test ./...

FROM deps AS build
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/monolith .


FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates tzdata vips \
    && adduser -D -H -u 10001 app \
    && mkdir -p /app/data && chown app:app /app/data

WORKDIR /app
COPY --from=build /out/monolith /app/monolith
USER app

ENV DATA_DIR=/app/data
VOLUME /app/data
EXPOSE 3000

HEALTHCHECK --interval=10s --timeout=3s --start-period=20s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:3000/healthz || exit 1

ENTRYPOINT ["/app/monolith"]
