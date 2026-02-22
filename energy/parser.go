package energy

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func parseCSV(r io.Reader) ([]Row, error) {
	reader := csv.NewReader(r)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}

	deviceNames, err := parseHeader(header)
	if err != nil {
		return nil, fmt.Errorf("parsing header: %w", err)
	}

	var rows []Row
	for {
		record, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("reading record: %w", err)
		}

		if len(record) == 0 {
			continue
		}
		if len(record) != len(header) {
			return nil, fmt.Errorf("unexpected record length: %d", len(record))
		}

		row, err := parseRow(record, deviceNames)
		if err != nil {
			return nil, fmt.Errorf("parsing record: %w", err)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func parseHeader(record []string) ([]string, error) {
	if len(record) < 2 {
		return nil, fmt.Errorf("header has too few fields: %d", len(record))
	}
	if record[0] != "timestamp" || record[1] != "Hz" {
		return nil, fmt.Errorf("unexpected header fields: %v", record[:2])
	}

	devices := make([]string, 0, (len(record)-2)/5)
	for i := 2; i < len(record); i += 5 {
		if i+4 >= len(record) {
			return nil, fmt.Errorf("header has too few fields for device: %d", len(record))
		}
		name, _, ok := strings.Cut(record[i], ".")
		if !ok {
			return nil, fmt.Errorf("unexpected header field for device: %s", record[i])
		}
		devices = append(devices, name)
	}
	return devices, nil
}

func parseRow(record, devices []string) (Row, error) {
	if len(record) < 2 {
		return Row{}, fmt.Errorf("record has too few fields: %d", len(record))
	}

	timestamp, err := strconv.ParseInt(record[0], 10, 64)
	if err != nil {
		return Row{}, fmt.Errorf("parsing timestamp: %w", err)
	}
	hz, err := strconv.ParseFloat(record[1], 64)
	if err != nil {
		return Row{}, fmt.Errorf("parsing Hz: %w", err)
	}

	row := Row{
		Timestamp: time.Unix(timestamp, 0),
		Hz:        hz,
		Devices:   make([]Device, 0, len(devices)),
	}
	for i, name := range devices {
		idx := 2 + i*5
		if idx+4 >= len(record) {
			return Row{}, fmt.Errorf("record has too few fields for device %s: %d", name, len(record))
		}
		volts, err := strconv.ParseFloat(record[idx], 64)
		if err != nil {
			return Row{}, fmt.Errorf("parsing volts for device %s: %w", name, err)
		}
		amps, err := strconv.ParseFloat(record[idx+1], 64)
		if err != nil {
			return Row{}, fmt.Errorf("parsing amps for device %s: %w", name, err)
		}
		watts, err := strconv.ParseFloat(record[idx+2], 64)
		if err != nil {
			return Row{}, fmt.Errorf("parsing watts for device %s: %w", name, err)
		}
		wattHours, err := strconv.ParseFloat(record[idx+3], 64)
		if err != nil {
			return Row{}, fmt.Errorf("parsing watt-hours for device %s: %w", name, err)
		}
		powerFactor, err := strconv.ParseFloat(record[idx+4], 64)
		if err != nil {
			return Row{}, fmt.Errorf("parsing power factor for device %s: %w", name, err)
		}

		row.Devices = append(row.Devices, Device{
			Name:        name,
			Volts:       volts,
			Amps:        amps,
			Watts:       watts,
			WattHours:   wattHours,
			PowerFactor: powerFactor,
		})
	}
	return row, nil
}
