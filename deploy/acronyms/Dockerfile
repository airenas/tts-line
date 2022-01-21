## encrypt file
FROM airenas/secure-file:0.1.2 as encryptor
WORKDIR /app
COPY data/acronyms.txt /app/
ARG SECRET
# use the value to set the ENV var default
ENV SECRET=$SECRET
RUN ./encrypt-file -f acronyms.txt -s $SECRET > acronyms.txt.aes


## final image
FROM iron/base:3.14.0

ARG BUILD_VERSION=0.1

LABEL org.opencontainers.image.version=$BUILD_VERSION \
      org.opencontainers.image.authors="airenass@gmail.com" \
      name="acronyms" \
      maintainer="airenass@gmail.com" \
      vendor="airenass@gmail.com" \
      version=$BUILD_VERSION \
      release=$BUILD_VERSION \
      summary="This image is used to host the Acronyms service" \
      description="This image is used to host the Acronyms service" 

WORKDIR /app

ENV UID=999

RUN addgroup app && adduser -S -G app --uid $UID app
RUN mkdir -p /app && chown -R app:app /app   

COPY --from=encryptor /app/check-decrypt-file /app/
COPY --from=encryptor /app/acronyms.txt.aes /app/
COPY data/acronyms /app/
COPY config.yaml /app/
COPY start.sh /app/

RUN chmod +x /app/start.sh
RUN chown app:app /app/* /app
 
USER app

ENTRYPOINT ["./start.sh"]
