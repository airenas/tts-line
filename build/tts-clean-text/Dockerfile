#########################################################################################
## docker will invoke this file from ../.. dir in order to access code
#########################################################################################
FROM golang:1.17-alpine as builder

ARG BUILD_VERSION=0.1

WORKDIR /go/src/
ENV CGO_ENABLED=0

COPY . /go/src
RUN CGO_ENABLED=0 go build -o /go/bin/tts-text-clean -ldflags "-X main.version=$BUILD_VERSION" cmd/tts-text-clean/main.go
#####################################################################################
FROM iron/base:3.14.0

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

ENV UID=999
EXPOSE 8000

RUN addgroup app && adduser -S -G app --uid $UID app
RUN mkdir -p /app && chown -R app:app /app   

COPY -from=builder d/go/bin/tts-text-clean /app/
COPY build/tts-clean-text/config.yaml /app/

RUN chown app:app /app/* /app
 
USER app

ENTRYPOINT ["./tts-text-clean"]
