# Build context: back/ (copies common/). cgo + libvips can't cross-compile like the other
# services, so every stage runs on the target platform (emulated with QEMU in the pipeline).
FROM golang:1.26.2-alpine AS deps
RUN apk add --no-cache gcc musl-dev pkgconf vips-dev
WORKDIR /src
COPY common/go.mod common/go.sum ./common/
COPY manga-worker/go.mod manga-worker/go.sum ./manga-worker/
WORKDIR /src/manga-worker
RUN go mod download
COPY common/ /src/common/
COPY manga-worker/ ./

# Unit tests with libvips: docker build --target test
FROM deps AS test
RUN go vet ./... && go test ./...

FROM deps AS build
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/manga-worker .


FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates tzdata vips && adduser -D -H -u 10001 app

WORKDIR /app
COPY --from=build /out/manga-worker /app/manga-worker
USER app

EXPOSE 3003

ENTRYPOINT ["/app/manga-worker"]
