// Package config provides structures and functions to parse YAML configuration for the relay.
package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the configuration for the relay, including the
// URL to listen on and the databases to forward data to.
type Config struct {
	URL       string     `yaml:"url"`
	InitialTS time.Time  `yaml:"initialTs"`
	Databases []Database `yaml:"databases"`
}

// Database is the configuration for a single database.
type Database struct {
	Type      string     `yaml:"type"`
	InfluxDB2 *InfluxDB2 `yaml:"influxdb2,omitempty"`
	InfluxDB3 *InfluxDB3 `yaml:"influxdb3,omitempty"`
}

// InfluxDB2 is the configuration for an InfluxDB 2.x database.
type InfluxDB2 struct {
	URL    string `yaml:"url"`
	Org    string `yaml:"org"`
	Bucket string `yaml:"bucket"`
	Token  string `yaml:"token"`
}

// InfluxDB3 is the configuration for an InfluxDB 3.x database.
type InfluxDB3 struct {
	URL      string `yaml:"url"`
	Database string `yaml:"database"`
	Token    string `yaml:"token"`
}

// Parse takes YAML data and unmarshals it into a Config struct.
func Parse(data []byte) (Config, error) {
	var cfg Config
	err := yaml.Unmarshal(data, &cfg)
	return cfg, err
}
