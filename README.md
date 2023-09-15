# Update Schema Description For Metrics

This processor (intended for use through `ocb`) will update a Honeycomb dataset
schema with metric descriptions. Experimental. Use at own risk.

## Building

1. Install the Collector builder `go install go.opentelemetry.io/collector/cmd/builder@latest`
2. Run `builder --config builder-config.yaml`
3. Run `./otelcol-mdp --config=collector-config.yaml` with appropriate values in
   api_key and dataset. Be sure that the dataset matches between your export and
   your processor config.