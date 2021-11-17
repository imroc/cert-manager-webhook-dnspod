FROM alpine:latest

COPY ./bin/webhook /usr/local/bin/webhook

ENTRYPOINT ["webhook"]
