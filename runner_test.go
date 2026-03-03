package relay_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hamba/cmd/v3/observe"
	"github.com/hamba/testutils/retry"
	relay "github.com/nrwiersma/aura-mon-relay"
	"github.com/nrwiersma/aura-mon-relay/database"
	"github.com/nrwiersma/aura-mon-relay/energy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testInterval = 5

func TestRunner_Run(t *testing.T) {
	obs := observe.NewFake()

	storedTS := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	rowTS := storedTS.Add(testInterval * time.Second)
	rows := []energy.Row{{
		Timestamp: rowTS,
		Hz:        60,
		Devices: []energy.Device{
			{Name: "dev1", Volts: 120, Amps: 1, Watts: 120, WattHours: 1.5, PowerFactor: 0.9},
			{Name: "dev2", Volts: 240, Amps: 2, Watts: 480, WattHours: 3.0, PowerFactor: 0.95},
		},
	}}

	storage := &mockStorage{}
	storage.Test(t)
	storage.On("Read").Return(storedTS, nil).Once()
	storage.On("Write", rowTS).Return(nil).Once()
	client := &mockClient{}
	client.Test(t)
	client.On("Get", storedTS.Add(testInterval*time.Second), testInterval).Return(rows, nil).Once()
	db := &mockDB{}
	db.Test(t)
	db.On("Write", mock.MatchedBy(func(metrics []database.Metric) bool {
		return len(metrics) == 2
	})).Return(nil).Once()

	runner := relay.NewRunner(client, []relay.DB{db}, storage, time.Now(), obs)

	go func() {
		err := runner.Run(t.Context())
		assert.NoError(t, err)
	}()

	retry.Run(t, func(t *retry.SubT) {
		storage.AssertExpectations(t)
		client.AssertExpectations(t)
		db.AssertExpectations(t)
	})
}

func TestRunner_RunStorageReadError(t *testing.T) {
	client := &mockClient{}
	storage := &mockStorage{}
	storage.Test(t)
	storage.On("Read").Return(time.Time{}, errors.New("read failed")).Once()

	runner := relay.NewRunner(client, nil, storage, time.Now(), observe.NewFake())

	err := runner.Run(t.Context())

	require.Error(t, err)
	require.ErrorContains(t, err, "reading last timestamp from storage")
	storage.AssertExpectations(t)
}

func TestRunner_RunInitialClientError(t *testing.T) {
	storage := &mockStorage{}
	storage.Test(t)
	storage.On("Read").Return(time.Time{}, nil).Once()
	client := &mockClient{}
	client.Test(t)
	client.On("Get", mock.Anything, testInterval).Return(nil, errors.New("client failed")).Once()

	runner := relay.NewRunner(client, []relay.DB{&mockDB{}}, storage, time.Now(), observe.NewFake())

	err := runner.Run(t.Context())

	require.Error(t, err)
	require.ErrorContains(t, err, "relaying metrics")
	require.ErrorContains(t, err, "getting metrics")
	storage.AssertExpectations(t)
	client.AssertExpectations(t)
}

func TestRunner_RunInitialDBWriteError(t *testing.T) {
	rowTS := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	rows := []energy.Row{{
		Timestamp: rowTS,
		Hz:        60,
		Devices: []energy.Device{
			{Name: "dev1", Volts: 120, Amps: 1, Watts: 120, WattHours: 1.5, PowerFactor: 0.9},
		},
	}}
	storage := &mockStorage{}
	storage.Test(t)
	storage.On("Read").Return(time.Time{}, nil).Once()
	client := &mockClient{}
	client.Test(t)
	client.On("Get", mock.Anything, testInterval).Return(rows, nil).Once()
	db := &mockDB{}
	db.Test(t)
	db.On("Write", mock.Anything).Return(errors.New("db down")).Once()

	runner := relay.NewRunner(client, []relay.DB{db}, storage, time.Now(), observe.NewFake())

	err := runner.Run(t.Context())

	require.Error(t, err)
	require.ErrorContains(t, err, "relaying metrics")
	require.ErrorContains(t, err, "sending metrics")
	require.ErrorContains(t, err, "writing to db")
	storage.AssertExpectations(t)
	client.AssertExpectations(t)
	db.AssertExpectations(t)
}

func TestRunner_RunInitialStorageWriteError(t *testing.T) {
	rowTS := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	rows := []energy.Row{{
		Timestamp: rowTS,
		Hz:        60,
		Devices: []energy.Device{
			{Name: "dev1", Volts: 120, Amps: 1, Watts: 120, WattHours: 1.5, PowerFactor: 0.9},
		},
	}}
	storage := &mockStorage{}
	storage.Test(t)
	storage.On("Read").Return(time.Time{}, nil).Once()
	storage.On("Write", rowTS).Return(errors.New("write failed")).Once()
	client := &mockClient{}
	client.Test(t)
	client.On("Get", mock.Anything, testInterval).Return(rows, nil).Once()
	db := &mockDB{}
	db.Test(t)
	db.On("Write", mock.Anything).Return(nil).Once()

	runner := relay.NewRunner(client, []relay.DB{db}, storage, time.Now(), observe.NewFake())

	err := runner.Run(t.Context())

	require.Error(t, err)
	require.ErrorContains(t, err, "relaying metrics")
	require.ErrorContains(t, err, "writing to storage")
	storage.AssertExpectations(t)
	client.AssertExpectations(t)
	db.AssertExpectations(t)
}

func TestRunner_RunEmptyRows(t *testing.T) {
	storage := &mockStorage{}
	storage.Test(t)
	storage.On("Read").Return(time.Time{}, nil).Once()
	client := &mockClient{}
	client.Test(t)
	client.On("Get", mock.Anything, testInterval).Return([]energy.Row{}, nil).Once()
	db := &mockDB{}

	runner := relay.NewRunner(client, []relay.DB{db}, storage, time.Now(), observe.NewFake())

	go func() {
		_ = runner.Run(t.Context())
	}()

	retry.Run(t, func(t *retry.SubT) {
		storage.AssertNotCalled(t, "Write", mock.Anything)
		storage.AssertExpectations(t)
		client.AssertExpectations(t)
		db.AssertNotCalled(t, "Write", mock.Anything)
	})
}

type mockClient struct {
	mock.Mock
}

func (m *mockClient) Get(_ context.Context, start time.Time, intvl int) ([]energy.Row, error) {
	args := m.Called(start, intvl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]energy.Row), args.Error(1)
}

type mockDB struct {
	mock.Mock
}

func (m *mockDB) Write(_ context.Context, metrics []database.Metric) error {
	args := m.Called(metrics)
	return args.Error(0)
}

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) Read() (time.Time, error) {
	args := m.Called()
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *mockStorage) Write(ts time.Time) error {
	args := m.Called(ts)
	return args.Error(0)
}
