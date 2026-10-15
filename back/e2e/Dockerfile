# Build context: back/ (copies common/ and the manga-worker samples)
# Runs in the cluster: /app/e2e.test in the sandbox namespace and /app/upload-samples in production.
FROM --platform=$BUILDPLATFORM golang:1.26.2-alpine AS build
ARG TARGETOS=linux
ARG TARGETARCH
WORKDIR /src
COPY common/go.mod common/go.sum ./common/
COPY e2e/go.mod e2e/go.sum ./e2e/
WORKDIR /src/e2e
RUN go mod download
COPY common/ /src/common/
COPY e2e/ ./
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH
RUN go test -tags e2e -c -o /out/e2e.test . \
    && go build -trimpath -ldflags="-s -w" -o /out/upload-samples ./cmd/upload-samples


FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates tzdata && adduser -D -H -u 10001 app

WORKDIR /app
COPY --from=build /out/ /app/
COPY manga-worker/samples/ /app/samples/
USER app

ENTRYPOINT ["/app/e2e.test", "-test.v", "-test.timeout=12m"]
