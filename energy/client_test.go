package energy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nrwiersma/aura-mon-relay/energy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testData = `timestamp,Hz,test1.V,test1.A,test1.W,test1.Wh,test1.PF,test2.V,test2.A,test2.W,test2.Wh,test2.PF
1771520870,50.10,227.035,0.200,18.381,0.025626,0.4056,226.871,0.000,0.000,0.000000,0.0000
1771520875,50.02,226.194,0.118,11.766,0.016502,0.4391,226.173,0.000,0.000,0.000000,0.0000
1771520880,49.95,226.381,0.127,8.580,0.011988,0.2984,226.218,0.000,0.000,0.000000,0.0000
`

func TestClient_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/energy", r.URL.Path)

		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte(testData))
	}))
	defer srv.Close()

	client, err := energy.NewClient(srv.URL)
	require.NoError(t, err)

	got, err := client.Get(t.Context(), time.Unix(1771520870, 0), 5)

	require.NoError(t, err)
	want := []energy.Row{
		{
			Timestamp: time.Unix(1771520870, 0),
			Hz:        50.10,
			Devices: []energy.Device{
				{Name: "test1", Volts: 227.035, Amps: 0.200, Watts: 18.381, WattHours: 0.025626, PowerFactor: 0.4056},
				{Name: "test2", Volts: 226.871, Amps: 0.000, Watts: 0.000, WattHours: 0.000000, PowerFactor: 0.0000},
			},
		},
		{
			Timestamp: time.Unix(1771520875, 0),
			Hz:        50.02,
			Devices: []energy.Device{
				{Name: "test1", Volts: 226.194, Amps: 0.118, Watts: 11.766, WattHours: 0.016502, PowerFactor: 0.4391},
				{Name: "test2", Volts: 226.173, Amps: 0.000, Watts: 0.000, WattHours: 0.000000, PowerFactor: 0.0000},
			},
		},
		{
			Timestamp: time.Unix(1771520880, 0),
			Hz:        49.95,
			Devices: []energy.Device{
				{Name: "test1", Volts: 226.381, Amps: 0.127, Watts: 8.580, WattHours: 0.011988, PowerFactor: 0.2984},
				{Name: "test2", Volts: 226.218, Amps: 0.000, Watts: 0.000, WattHours: 0.000000, PowerFactor: 0.0000},
			},
		},
	}
	assert.Equal(t, want, got)
}
