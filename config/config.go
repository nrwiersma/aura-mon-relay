// Package config parses YAML configuration for the relay.
package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level relay configuration.
type Config struct {
	URL       string     `yaml:"url"`
	InitialTS time.Time  `yaml:"initialTs"` // Timestamp to start from when no stored state exists.
	Databases []Database `yaml:"databases"`
}

// Database describes a single destination database and its connection settings.
type Database struct {
	Type      string     `yaml:"type"`
	InfluxDB2 *InfluxDB2 `yaml:"influxdb2,omitempty"`
	InfluxDB3 *InfluxDB3 `yaml:"influxdb3,omitempty"`
}

// InfluxDB2 holds connection settings for an InfluxDB 2.x database.
type InfluxDB2 struct {
	URL    string `yaml:"url"`
	Org    string `yaml:"org"`
	Bucket string `yaml:"bucket"`
	Token  string `yaml:"token"`
}

// InfluxDB3 holds connection settings for an InfluxDB 3.x database.
type InfluxDB3 struct {
	URL      string `yaml:"url"`
	Database string `yaml:"database"`
	Token    string `yaml:"token"`
}

// Parse unmarshals YAML data into a Config.
func Parse(data []byte) (Config, error) {
	var cfg Config
	err := yaml.Unmarshal(data, &cfg)
	return cfg, err
}
