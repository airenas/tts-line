FROM iron/base
LABEL Description="This image is used to host the TTS Clean Text service" \ 
      Maintainer="airenass@gmail.com" Vendor="Airenas Vaiciunas" Version="1.0" 

WORKDIR /app

COPY data/tts-text-clean /app/
COPY config.yaml /app/

ENTRYPOINT ["./tts-clean-text"]
