# Acuvity OpenTelemetry Collector exporter

This is an OpenTelemetry Collector compatible go module that can be used together with the [OpenTelemetry Collector Builder](https://opentelemetry.io/docs/collector/custom-collector/) ([repo](https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder)) to be able to export traces to the Acuvity platform.

## Build Instructions

Familiarity with the [OpenTelemetry Collector Builder](https://opentelemetry.io/docs/collector/custom-collector/) is assumed in this section.

Create a `builder-config.yaml` file with at least the following contents; obviously ensure to include this golang module in the exporters section. You can include any other module here that you need of course:

```yaml
dist:
  name: otelcol-acuvity
  description: OTel Collector distribution including the Acuvity exporter
  output_path: ./otelcol-acuvity

exporters:
  - gomod: go.opentelemetry.io/collector/exporter/otlpexporter v0.127.0
  - gomod: go.opentelemetry.io/collector/exporter/debugexporter v0.127.0
  - gomod: github.com/acuvity/otelexporter v0.1.0

processors:
  - gomod:
      go.opentelemetry.io/collector/processor/batchprocessor v0.127.0

receivers:
  - gomod:
      go.opentelemetry.io/collector/receiver/otlpreceiver v0.127.0

providers:
  - gomod: go.opentelemetry.io/collector/confmap/provider/envprovider v1.18.0
  - gomod: go.opentelemetry.io/collector/confmap/provider/fileprovider v1.18.0
  - gomod: go.opentelemetry.io/collector/confmap/provider/httpprovider v1.18.0
  - gomod: go.opentelemetry.io/collector/confmap/provider/httpsprovider v1.18.0
  - gomod: go.opentelemetry.io/collector/confmap/provider/yamlprovider v1.18.0
```

Then generate the sources by running: `ocb --config=builder-config.yaml`

Your own generated OpenTelemetry Collector will then be available at `./otelcol-acuvity`

## Development

The above is conveniently prepared within a `builder-config.yaml` file which is in this repo which is perfect for development.
Copy the `otel-config-template.yaml` to `otel-config.yaml` with a simple `cp otel-config{-template,}.yaml`.
Then edit the `otel-config.yaml` to add your App Token to the configuration.

For building the collector run:

```shell
make build
```

And for running the built collector using the `otel-config.yaml`, simply run:

```shell
make run
```

You can then generate traces and feed it into the collector by running:

```shell
make generate-traces
```

For more Makefile targets run `make help`.
