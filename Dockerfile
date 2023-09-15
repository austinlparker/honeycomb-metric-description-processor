# Builder stage
FROM golang:1.21.1 as builder
WORKDIR /app

RUN go install go.opentelemetry.io/collector/cmd/builder@latest

COPY builder-config-k8s.yaml ./builder-config.yaml
COPY metricdescriptionprocessor ./metricdescriptionprocessor

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN builder --config builder-config.yaml

COPY collector-config.local.yaml /etc/otelcol/config.yaml

ENTRYPOINT ["/tmp/dist/otelcol-custom"]
CMD ["--config", "/etc/otelcol/config.yaml"]
EXPOSE 4317 4318