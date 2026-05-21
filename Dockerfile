FROM alpine:3.20

RUN addgroup -S orbit && adduser -S orbit -G orbit

LABEL org.opencontainers.image.source="https://github.com/erdembestas/orbit"
LABEL org.opencontainers.image.description="Orbit API and controller image for the single-cluster Kubernetes control plane preview."
LABEL org.opencontainers.image.licenses="Apache-2.0"

USER orbit
WORKDIR /app

COPY orbit-api /usr/local/bin/orbit-api
COPY orbit-controller /usr/local/bin/orbit-controller

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/orbit-api"]
