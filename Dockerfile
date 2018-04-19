FROM alpine:3.6

RUN apk update --no-cache && apk add ca-certificates

COPY kctlr-docker-auth /kctlr-docker-auth

ENTRYPOINT ["/kctlr-docker-auth"]
