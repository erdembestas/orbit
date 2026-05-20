FROM alpine:3.20

RUN addgroup -S orbit && adduser -S orbit -G orbit

USER orbit
WORKDIR /app

COPY orbit-api /usr/local/bin/orbit-api
COPY orbit-controller /usr/local/bin/orbit-controller

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/orbit-api"]
