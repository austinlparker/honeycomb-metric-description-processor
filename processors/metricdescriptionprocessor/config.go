package metricdescriptionprocessor

// This struct holds the configuration for our processor.
type Config struct {
	APIKey   string `mapstructure:"api_key"`
	Dataset  string `mapstructure:"dataset"`
}