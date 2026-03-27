package database_test

import (
	"context"
	"errors"
	"testing"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/nrwiersma/aura-mon-relay/database"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewInfluxDB2_ValidatesURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "valid URL",
			rawURL:  "http://localhost:8086",
			wantErr: require.NoError,
		},
		{
			name:    "invalid URL",
			rawURL:  "://broken",
			wantErr: require.Error,
		},
		{
			name:    "missing host",
			rawURL:  "http://",
			wantErr: require.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := database.NewInfluxDB2(test.rawURL, "token", "org", "bucket")

			test.wantErr(t, err)
		})
	}
}

func TestInfluxDB2Write_EmptyMetricsNoOp(t *testing.T) {
	t.Parallel()

	writer := &mockInfluxDB2Writer{}
	writer.Test(t)

	client := &fakeInfluxDB2Client{writer: writer}

	db := database.NewInfluxDB2WithClient(client, "org", "bucket")

	err := db.Write(t.Context(), nil)

	require.NoError(t, err)
	writer.AssertNotCalled(t, "WritePoint", mock.Anything)
}

func TestInfluxDB2Write_WritesPoints(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	writer := &mockInfluxDB2Writer{}
	writer.Test(t)
	writer.On("WritePoint", []*write.Point{
		write.NewPoint(
			"meter",
			map[string]string{"device": "main"},
			map[string]any{"watts": 123.4},
			now.Truncate(time.Second),
		),
	}).Return(nil).Once()
	client := &fakeInfluxDB2Client{writer: writer}

	db := database.NewInfluxDB2WithClient(client, "org", "bucket")

	err := db.Write(t.Context(), []database.Metric{{
		Measurement: "meter",
		Timestamp:   now.Unix(),
		Tags:        map[string]string{"device": "main"},
		Fields:      map[string]float64{"watts": 123.4},
	}})

	require.NoError(t, err)
	writer.AssertExpectations(t)
}

func TestInfluxDB2Write_PropagatesWriteError(t *testing.T) {
	t.Parallel()

	writer := &mockInfluxDB2Writer{}
	writer.Test(t)
	writer.On("WritePoint", mock.Anything).Return(errors.New("test")).Once()
	client := &fakeInfluxDB2Client{writer: writer}

	db := database.NewInfluxDB2WithClient(client, "org", "bucket")

	err := db.Write(t.Context(), []database.Metric{{Measurement: "meter", Timestamp: 1}})

	require.Error(t, err)
	require.ErrorContains(t, err, "writing metrics to influxdb2")
}

type fakeInfluxDB2Client struct {
	influxdb2.Client

	writer *mockInfluxDB2Writer
}

func (c *fakeInfluxDB2Client) WriteAPIBlocking(string, string) api.WriteAPIBlocking {
	return c.writer
}

type mockInfluxDB2Writer struct {
	api.WriteAPIBlocking
	mock.Mock
}

func (w *mockInfluxDB2Writer) WritePoint(_ context.Context, points ...*write.Point) error {
	args := w.Called(points)
	return args.Error(0)
}
