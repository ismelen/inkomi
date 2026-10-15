# Build context: back/ (copies common/). Multi-arch: Go cross-compiles for TARGETARCH on the
# builder's own platform, so no emulation is needed.
FROM --platform=$BUILDPLATFORM golang:1.26.2-alpine AS deps
WORKDIR /src
COPY common/go.mod common/go.sum ./common/
COPY transaction-manager/go.mod transaction-manager/go.sum ./transaction-manager/
WORKDIR /src/transaction-manager
RUN go mod download
COPY common/ /src/common/
COPY transaction-manager/ ./

# Unit tests: docker build --target test
FROM deps AS test
RUN go vet ./... && go test ./...

FROM deps AS build
ARG TARGETOS=linux
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/transaction-manager .


FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates tzdata && adduser -D -H -u 10001 app

WORKDIR /app
COPY --from=build /out/transaction-manager /app/transaction-manager
COPY --from=build /src/transaction-manager/migrations /app/migrations
USER app

ENTRYPOINT ["/app/transaction-manager"]
