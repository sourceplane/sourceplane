FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY sp /usr/local/bin/sp

ENTRYPOINT ["/usr/local/bin/sp"]
