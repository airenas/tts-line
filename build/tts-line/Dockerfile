#########################################################################################
## docker will invoke this file from ../.. dir in order to access code
#########################################################################################
FROM golang:1.21-alpine AS builder

ARG BUILD_VERSION=0.1

WORKDIR /go/src/
ENV CGO_ENABLED=0

COPY . /go/src

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o /go/bin/tts-line -trimpath -ldflags "-s -w -X main.version=$BUILD_VERSION" cmd/tts-line/main.go
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o /go/bin/tts-export -trimpath  -ldflags "-s -w -X main.version=$BUILD_VERSION" cmd/tts-export/main.go
#####################################################################################
FROM alpine:3.15 AS runner

ARG BUILD_VERSION=0.1

LABEL org.opencontainers.image.version=$BUILD_VERSION \
      org.opencontainers.image.authors="airenass@gmail.com" \
      name="tts-line" \
      maintainer="airenass@gmail.com" \
      vendor="airenass@gmail.com" \
      version=$BUILD_VERSION \
      release=$BUILD_VERSION \
      summary="This image is used to host the TTS Line service" \
      description="This image is used to host the TTS Line service" 

WORKDIR /app

ENV UID=999
EXPOSE 8000

RUN addgroup app && adduser -S -G app app --uid $UID app

RUN mkdir -p /app && chown -R app:app /app
RUN mkdir -p /suffixes && chown -R app:app /suffixes      

COPY --from=builder /go/bin/tts-line /app/
COPY --from=builder /go/bin/tts-export /app/
COPY build/tts-line/config.yaml /app/

RUN chown app:app /app/* /app
 
USER app

ENTRYPOINT ["./tts-line"]
