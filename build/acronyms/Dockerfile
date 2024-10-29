#########################################################################################
## docker will invoke this file from ../.. dir in order to access code
#########################################################################################
FROM golang:1.21-alpine AS builder

ARG BUILD_VERSION=0.1

WORKDIR /go/src/
ENV CGO_ENABLED=0

COPY . /go/src

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o /go/bin/acronyms -trimpath -ldflags "-s -w -X main.version=$BUILD_VERSION" cmd/acronyms/main.go

#########################################################################################
## encrypt file
FROM airenas/secure-file:0.1.2 AS encryptor
WORKDIR /app
COPY build/acronyms/data/acronyms.txt /app/
ARG SECRET
# use the value to set the ENV var default
ENV SECRET=$SECRET
RUN ./encrypt-file -f acronyms.txt -s $SECRET > acronyms.txt.aes

#########################################################################################
## final image
FROM alpine:3.15 AS runner

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
EXPOSE 8000

RUN addgroup app && adduser -S -G app app --uid $UID app

RUN mkdir -p /app && chown -R app:app /app

COPY --from=encryptor /app/check-decrypt-file /app/
COPY --from=encryptor /app/acronyms.txt.aes /app/
COPY --from=builder /go/bin/acronyms /app/
COPY build/acronyms/config.yaml /app/
COPY build/acronyms/start.sh /app/

RUN chown app:app /app/* /app
 
USER app

ENTRYPOINT ["./start.sh"]
