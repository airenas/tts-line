#########################################################################################
## docker will invoke this file from ../.. dir in order to access code
#########################################################################################
FROM golang:1.22-alpine AS builder

ARG BUILD_VERSION=0.1

WORKDIR /go/src/
ENV CGO_ENABLED=0

COPY . /go/src

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o /go/bin/tts-text-clean -trimpath  -ldflags "-s -w -X main.version=$BUILD_VERSION" cmd/tts-text-clean/main.go
#####################################################################################
FROM gcr.io/distroless/static-debian12 AS runner

ARG BUILD_VERSION=0.1

LABEL org.opencontainers.image.version=$BUILD_VERSION \
      org.opencontainers.image.authors="airenass@gmail.com" \
      name="tts-text-clean" \
      maintainer="airenass@gmail.com" \
      vendor="airenass@gmail.com" \
      version=$BUILD_VERSION \
      release=$BUILD_VERSION \
      summary="This image is used to host the TTS Clean Text service" \
      description="This image is used to host the TTS Clean Text service" 

WORKDIR /app

EXPOSE 8000

COPY --from=builder /go/bin/tts-text-clean /app/
COPY build/tts-text-clean/config.yaml /app/

ENTRYPOINT ["./tts-text-clean"]
