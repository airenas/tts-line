#########################################################################################
## docker will invoke this file from ../.. dir in order to access code
#########################################################################################
FROM golang:1.25-alpine AS builder

ARG BUILD_VERSION=0.1

WORKDIR /go/src/
ENV CGO_ENABLED=0

COPY . /go/src

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o /go/bin/clitics -trimpath -ldflags "-s -w -X main.version=$BUILD_VERSION" cmd/clitics/main.go
#########################################################################################
## data
#########################################################################################
FROM airenas/clitics:0.1.206 AS data
#########################################################################################
## encrypt file
#########################################################################################
FROM airenas/secure-file:0.1.2 AS encryptor
WORKDIR /app
ARG SECRET
ARG SECRET_FROM

# use the value to set the ENV var default
ENV SECRET_FROM=$SECRET_FROM
ENV SECRET=$SECRET
COPY --from=data /app/phrases.txt.aes /app/
COPY --from=data /app/clitics.txt /app/

RUN ./check-decrypt-file -f phrases.txt -s $SECRET_FROM

RUN ./encrypt-file -f phrases.txt -s $SECRET > phrases.txt.aes
#########################################################################################
## final image
#########################################################################################
FROM alpine:3.22 AS runner

ARG BUILD_VERSION=0.1

LABEL org.opencontainers.image.version=$BUILD_VERSION \
      org.opencontainers.image.authors="airenass@gmail.com" \
      name="tts-line" \
      maintainer="airenass@gmail.com" \
      vendor="airenass@gmail.com" \
      version=$BUILD_VERSION \
      release=$BUILD_VERSION \
      summary="This image is used to host clitics service" \
      description="This image is used to host clitics service" 

COPY LICENSE /LICENSE

WORKDIR /app

COPY --from=encryptor /app/check-decrypt-file /app/
COPY --from=encryptor /app/phrases.txt.aes /app/
COPY --from=encryptor /app/clitics.txt /app/

COPY --from=builder /go/bin/clitics /app/

COPY ./build/clitics/config.yaml /app/
COPY ./build/clitics/start.sh /app/

RUN chmod +x /app/start.sh

ENTRYPOINT ["./start.sh"]
