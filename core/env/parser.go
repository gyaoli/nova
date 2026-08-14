package env

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Load parses and validates one YAML service configuration without changing
// the active configuration. Unknown fields are rejected to catch typos early.
func Load(configFile string) (NodeConfig, error) {
	configFile = strings.TrimSpace(configFile)
	if configFile == "" {
		return NodeConfig{}, fmt.Errorf("config file cannot be empty")
	}

	v := viper.New()
	v.SetConfigFile(filepath.Clean(configFile))
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return NodeConfig{}, fmt.Errorf("read config %q: %w", configFile, err)
	}

	var config NodeConfig
	if err := v.UnmarshalExact(&config); err != nil {
		return NodeConfig{}, fmt.Errorf("decode config %q: %w", configFile, err)
	}
	if err := config.validate(); err != nil {
		return NodeConfig{}, fmt.Errorf("validate config %q: %w", configFile, err)
	}
	return config, nil
}
