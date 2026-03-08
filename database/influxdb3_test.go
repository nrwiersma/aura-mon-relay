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
	writer.On("WritePoints", mock.Anything, mock.Anything).Return(nil).Once()

	db := database.NewInfluxDB3WithClient(writer, "database")

	err := db.Write(t.Context(), nil)

	require.NoError(t, err)
	writer.AssertNotCalled(t, "WritePoints", mock.Anything, mock.Anything)
}

func TestInfluxDB3Write_WritesPoints(t *testing.T) {
	t.Parallel()

	writer := &mockInfluxDB3Client{}
	writer.Test(t)
	writer.On("WritePoints", mock.Anything, mock.Anything).Return(nil).Once()

	db := database.NewInfluxDB3WithClient(writer, "database")

	err := db.Write(t.Context(), []database.Metric{{
		Measurement: "meter",
		Timestamp:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
		Tags:        map[string]string{"device": "main"},
		Fields:      map[string]float64{"watts": 123.4},
	}})

	require.NoError(t, err)
	writer.AssertExpectations(t)
}

func TestInfluxDB3Write_PropagatesWriteError(t *testing.T) {
	t.Parallel()

	writer := &mockInfluxDB3Client{}
	writer.Test(t)
	writer.On("WritePoints", mock.Anything, mock.Anything).Return(errors.New("test")).Once()

	db := database.NewInfluxDB3WithClient(writer, "database")

	err := db.Write(t.Context(), []database.Metric{{Measurement: "meter", Timestamp: 1}})

	require.Error(t, err)
	require.ErrorContains(t, err, "writing metrics to influxdb3")
}

type mockInfluxDB3Client struct {
	mock.Mock
}

func (w *mockInfluxDB3Client) WritePoints(_ context.Context, _ []*influxdb3.Point, _ ...influxdb3.WriteOption) error {
	args := w.Called(mock.Anything, mock.Anything)
	return args.Error(0)
}

func (w *mockInfluxDB3Client) Close() error {
	args := w.Called()
	return args.Error(0)
}
