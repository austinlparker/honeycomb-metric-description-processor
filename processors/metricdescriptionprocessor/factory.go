package metricdescriptionprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

var processorCapabilities = consumer.Capabilities{MutatesData: false}

func NewFactory() processor.Factory {
	return processor.NewFactory(
		"attributes",
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, component.StabilityLevelBeta))
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createMetricsProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	return processorhelper.NewMetricsProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		newMetricsDescriptionProcessor(set.Logger, cfg).processMetrics,
		processorhelper.WithCapabilities(processorCapabilities))
}
