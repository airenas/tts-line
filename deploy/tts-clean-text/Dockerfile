FROM iron/base:3.14.0

ARG BUILD_VERSION=0.1

LABEL org.opencontainers.image.version=$BUILD_VERSION \
      org.opencontainers.image.authors="airenass@gmail.com" \
      name="big-tts-upload" \
      maintainer="airenass@gmail.com" \
      vendor="airenass@gmail.com" \
      version=$BUILD_VERSION \
      release=$BUILD_VERSION \
      summary="This image is used to host the TTS Clean Text service" \
      description="This image is used to host the TTS Clean Text service" 

WORKDIR /app

ENV UID=999

RUN addgroup app && adduser -S -G app --uid $UID app
RUN mkdir -p /app && chown -R app:app /app   

COPY data/tts-text-clean /app/
COPY config.yaml /app/

RUN chown app:app /app/* /app
 
USER app

ENTRYPOINT ["./tts-text-clean"]
