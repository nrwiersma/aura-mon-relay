package config_test

import (
	"testing"
	"time"

	"github.com/nrwiersma/aura-mon-relay/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    config.Config
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "allows valid config",
			yaml: `url: http://localhost:8080
initialTs: 2024-01-02T03:04:05Z
databases:
    - type: influxdb2
      influxdb2:
        url: http://localhost:8086
        org: acme
        bucket: telemetry
        token: t1
    - type: influxdb3
      influxdb3:
        url: https://cloud.example.com
        database: telemetry
        token: t2
`,
			want: config.Config{
				URL:       "http://localhost:8080",
				InitialTS: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
				Databases: []config.Database{
					{
						Type: "influxdb2",
						InfluxDB2: &config.InfluxDB2{
							URL:    "http://localhost:8086",
							Org:    "acme",
							Bucket: "telemetry",
							Token:  "t1",
						},
					},
					{
						Type: "influxdb3",
						InfluxDB3: &config.InfluxDB3{
							URL:      "https://cloud.example.com",
							Database: "telemetry",
							Token:    "t2",
						},
					},
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "allows url only",
			yaml: `url: http://localhost:9090
`,
			want: config.Config{
				URL:       "http://localhost:9090",
				Databases: nil,
			},
			wantErr: require.NoError,
		},
		{
			name:    "handles invalid yaml",
			yaml:    "url: [",
			wantErr: require.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := config.Parse([]byte(test.yaml))

			test.wantErr(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}
