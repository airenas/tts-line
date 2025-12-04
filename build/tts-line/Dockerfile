#########################################################################################
## docker will invoke this file from ../.. dir in order to access code
#########################################################################################
FROM golang:1.25-alpine AS builder

ARG BUILD_VERSION=0.1

WORKDIR /go/src/

RUN apk add build-base cmake git pkgconfig
RUN git clone --branch v1.2.6 --single-branch https://github.com/jiixyj/libebur128.git \
    && mkdir -p libebur128/build \
    && cd libebur128/build \
    && cmake .. -DBUILD_SHARED_LIBS=OFF -DCMAKE_POLICY_VERSION_MINIMUM=3.5 \
    && make \
    && make install

ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-I/usr/local/include"
ENV CGO_LDFLAGS="-L/usr/local/lib -lebur128"


COPY . /go/src

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o /go/bin/tts-line -trimpath -ldflags "-s -w -X main.version=$BUILD_VERSION" cmd/tts-line/main.go
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o /go/bin/tts-export -trimpath  -ldflags "-s -w -X main.version=$BUILD_VERSION" cmd/tts-export/main.go
#####################################################################################
FROM alpine:3.22 AS runner

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

COPY Licenses /Licenses
COPY LICENSE /LICENSE

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
