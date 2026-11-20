# Build context: back/api/ (the e2e tests are in the api module; the samples come from samples/).
# Runs from the compose e2e profile: /app/e2e.test against a throwaway stack and /app/upload-samples
# before production.
FROM --platform=$BUILDPLATFORM golang:1.26.2-alpine AS build
ARG TARGETOS=linux
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH
RUN go test -tags e2e -c -o /out/e2e.test ./e2e \
    && go build -trimpath -ldflags="-s -w" -o /out/upload-samples ./e2e/cmd/upload-samples


FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates tzdata && adduser -D -H -u 10001 app

WORKDIR /app
COPY --from=build /out/ /app/
COPY samples/*.jpg /app/samples/
USER app

ENTRYPOINT ["/app/e2e.test", "-test.v", "-test.timeout=12m"]
