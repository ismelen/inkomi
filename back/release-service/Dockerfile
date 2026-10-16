# Build context: back/ (like every service, although this one doesn't use common/). Multi-arch:
# Go cross-compiles for TARGETARCH on the builder's own platform, so no emulation is needed.
FROM --platform=$BUILDPLATFORM golang:1.26.2-alpine AS deps
WORKDIR /src/release-service
COPY release-service/go.mod release-service/go.sum ./
RUN go mod download
COPY release-service/ ./

# Unit tests: docker build --target test
FROM deps AS test
RUN go vet ./... && go test ./...

FROM deps AS build
ARG TARGETOS=linux
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/release-service .


FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates tzdata && adduser -D -H -u 10001 app

WORKDIR /app
COPY --from=build /out/release-service /app/release-service
USER app

ENTRYPOINT ["/app/release-service"]
