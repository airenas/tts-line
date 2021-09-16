FROM iron/base

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

RUN addgroup app && adduser -S -G app app

RUN mkdir -p /app && chown -R app:app /app   

COPY data/tts-line /app/
COPY data/tts-export /app/
COPY config.yaml /app/

RUN chown app:app /app/* /app
 
USER app

ENTRYPOINT ["./tts-line"]
