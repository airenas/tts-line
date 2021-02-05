## encrypt file
FROM airenas/secure-file:0.1.2 as encryptor
WORKDIR /app
COPY data/acronyms.txt /app/
ARG SECRET
# use the value to set the ENV var default
ENV SECRET=$SECRET
RUN ./encrypt-file -f acronyms.txt -s $SECRET > acronyms.txt.aes


## final image
FROM iron/base
LABEL Description="This image is used to host acronyms service" \ 
      Maintainer="airenass@gmail.com" Vendor="Airenas Vaiciunas" Version="1.0" 

WORKDIR /app

COPY --from=encryptor /app/check-decrypt-file /app/
COPY --from=encryptor /app/acronyms.txt.aes /app/
COPY data/acronyms /app/
COPY config.yaml /app/
COPY start.sh /app/
RUN ["chmod", "+x", "/app/start.sh"]

ENTRYPOINT ["./start.sh"]
