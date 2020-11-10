FROM iron/base
LABEL Description="This image is used to host the TTS Line service" \ 
      Maintainer="airenass@gmail.com" Vendor="Airenas Vaiciunas" Version="1.0" 

WORKDIR /app

COPY data/tts-line /app/
COPY config.yaml /app/

ENTRYPOINT ["./tts-line"]
