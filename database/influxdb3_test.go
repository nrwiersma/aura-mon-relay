package database_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/influxdb3"
	"github.com/nrwiersma/aura-mon-relay/database"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewInfluxDB3_ValidatesURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "invalid URL",
			rawURL:  "://broken",
			wantErr: require.Error,
		},
		{
			name:    "missing host",
			rawURL:  "https://",
			wantErr: require.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := database.NewInfluxDB3(test.rawURL, "token", "database")

			test.wantErr(t, err)
		})
	}
}

func TestInfluxDB3Write_EmptyMetricsNoOp(t *testing.T) {
	t.Parallel()

	writer := &mockInfluxDB3Client{}
	writer.Test(t)

	db := database.NewInfluxDB3WithClient(writer)

	err := db.Write(t.Context(), nil)

	require.NoError(t, err)
	writer.AssertNotCalled(t, "WritePoints", mock.Anything)
}

func TestInfluxDB3Write_WritesPoints(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	writer := &mockInfluxDB3Client{}
	writer.Test(t)
	writer.On("WritePoints", []*influxdb3.Point{
		{Values: &influxdb3.PointValues{
			MeasurementName: "meter",
			Tags:            map[string]string{"device": "main"},
			Fields:          map[string]any{"watts": 123.4},
			Timestamp:       now.Truncate(time.Second),
		}},
	}).Return(nil).Once()

	db := database.NewInfluxDB3WithClient(writer)

	err := db.Write(t.Context(), []database.Metric{{
		Measurement: "meter",
		Timestamp:   now.Unix(),
		Tags:        map[string]string{"device": "main"},
		Fields:      map[string]float64{"watts": 123.4},
	}})

	require.NoError(t, err)
	writer.AssertExpectations(t)
}

func TestInfluxDB3Write_PropagatesWriteError(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	writer := &mockInfluxDB3Client{}
	writer.Test(t)
	writer.On("WritePoints", []*influxdb3.Point{
		{Values: &influxdb3.PointValues{
			MeasurementName: "meter",
			Tags:            map[string]string{},
			Fields:          map[string]any{},
			Timestamp:       now.Truncate(time.Second),
		}},
	}).Return(errors.New("test")).Once()

	db := database.NewInfluxDB3WithClient(writer)

	err := db.Write(t.Context(), []database.Metric{{Measurement: "meter", Timestamp: now.Unix()}})

	require.Error(t, err)
	require.ErrorContains(t, err, "writing metrics to influxdb3")
	writer.AssertExpectations(t)
}

type mockInfluxDB3Client struct {
	mock.Mock
}

func (w *mockInfluxDB3Client) WritePoints(_ context.Context, pts []*influxdb3.Point, _ ...influxdb3.WriteOption) error {
	args := w.Called(pts)
	return args.Error(0)
}

func (w *mockInfluxDB3Client) Close() error {
	return nil
}
