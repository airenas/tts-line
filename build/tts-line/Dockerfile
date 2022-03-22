#########################################################################################
## docker will invoke this file from ../.. dir in order to access code
#########################################################################################
FROM golang:1.17-alpine as builder

ARG BUILD_VERSION=0.1

WORKDIR /go/src/
ENV CGO_ENABLED=0

# precache go.mod dependencies
COPY go.mod /go/src/
COPY go.sum /go/src/
RUN go mod download

# copy src
COPY . /go/src

RUN CGO_ENABLED=0 go build -o /go/bin/tts-line -ldflags "-X main.version=$BUILD_VERSION" cmd/tts-line/main.go
RUN CGO_ENABLED=0 go build -o /go/bin/tts-export -ldflags "-X main.version=$BUILD_VERSION" cmd/tts-export/main.go
#####################################################################################
FROM iron/base:3.14

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

COPY --from=builder /go/bin/tts-line /app/
COPY --from=builder /go/bin/tts-export /app/
COPY build/tts-line/config.yaml /app/

RUN chown app:app /app/* /app
 
USER app

ENTRYPOINT ["./tts-line"]
