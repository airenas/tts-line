## encrypt file
FROM airenas/secure-file:0.1.2 AS encryptor
WORKDIR /app
COPY data/phrases.txt /app/
ARG SECRET
# use the value to set the ENV var default
ENV SECRET=$SECRET
RUN ./encrypt-file -f phrases.txt -s $SECRET > phrases.txt.aes


## final image
FROM iron/base
LABEL Description="This image is used to host clitics service" \ 
      Maintainer="airenass@gmail.com" Vendor="Airenas Vaiciunas" Version="1.0" 

WORKDIR /app

COPY --from=encryptor /app/check-decrypt-file /app/
COPY --from=encryptor /app/phrases.txt.aes /app/
COPY data/clitics /app/
COPY data/clitics.txt /app/
COPY config.yaml /app/
COPY start.sh /app/
RUN ["chmod", "+x", "/app/start.sh"]

ENTRYPOINT ["./start.sh"]
