# aura-mon-relay

`aura-mon-relay` polls an Aura-Mon device for energy metrics and relays them into one or more time-series databases.

It currently supports:
- InfluxDB 2.x
- InfluxDB 3.x

## How It Works

- Fetches CSV energy data from the Aura-Mon `energy` endpoint every 5 seconds.
- Converts each row into per-device metrics (`hz`, `volts`, `amps`, `watts`, `watt_hours`, `power_factor`).
- Writes the metrics to all configured databases.
- Persists the last successful timestamp to a state file so restarts continue where they left off.

## Requirements

- Go `1.26.1+` (for building from source)
- Network access to your Aura-Mon instance
- At least one configured database destination (InfluxDB 2.x and/or 3.x)

## Installation

### Option 1: Build From Source (recommended for development)

```bash
git clone https://github.com/nrwiersma/aura-mon-relay.git
cd aura-mon-relay
go build -o relay ./cmd/relay
```

### Option 2: Install With `go install`

```bash
go install github.com/nrwiersma/aura-mon-relay/cmd/relay@latest
```

### Option 3: Release Artifacts

This project is configured to publish:
- Linux binaries (`amd64`, `arm64`)
- Docker images (`ghcr.io/nrwiersma/aura-mon-relay:<tag>`)
- Linux packages (`.deb`, `.rpm`)

If you use systemd packages, a unit file is included at `release/systemd/aura-mon-relay.service`.

## Configuration

Create a YAML config file (default: `config.yaml`):

```yaml
url: http://aura-mon.local
initialTs: 2024-01-01T00:00:00Z
databases:
  - type: influxdb2
    influxdb2:
      url: http://localhost:8086
      org: home
      bucket: auramon
      token: YOUR_INFLUXDB2_TOKEN
  - type: influxdb3
    influxdb3:
      url: https://your-influxdb3-host
      database: auramon
      token: YOUR_INFLUXDB3_TOKEN
```

Notes:
- `url` is required.
- `databases` must contain at least one destination.
- `initialTs` is optional. If omitted, relay starts from approximately `now - 1 minute`.

## Usage

### CLI

```bash
relay --config ./config.yaml --state ./relay.state
```

Important flags:
- `--config`, `-c`: path to YAML config (default: `config.yaml`)
- `--state`: path to state file (required)

Environment variable equivalents:
- `CONFIG`
- `STATE`
- `LOG_LEVEL`, `LOG_FORMAT`, `LOG_CTX`

Example with environment variables:

```bash
CONFIG=./config.yaml
STATE=./relay.state
LOG_LEVEL=info
relay
```

### Docker

The container expects the binary entrypoint and still requires config/state flags or env vars:

```bash
docker run --rm \
  -v "$(pwd)/config.yaml:/config.yaml:ro" \
  -v "$(pwd)/relay.state:/relay.state" \
  ghcr.io/nrwiersma/aura-mon-relay:<tag> \
  --config /config.yaml --state /relay.state
```

### Local InfluxDB (optional)

A local InfluxDB can be started with:

```bash
docker compose up -d
```

Default compose values expose InfluxDB at `http://localhost:8086`.

### systemd

An example systemd unit is provided in `release/systemd/aura-mon-relay.service` with:
- config file path: `/etc/aura-mon/relay.yaml`
- state file path: `/opt/aura-mon/relay.state`

Adjust paths and permissions for your environment.

## Development

Common make targets:

```bash
make fmt
make tidy
make lint
make test
make build
```

- `make build` uses GoReleaser snapshot mode.
- `make test` runs `go test -cover -race ./...`.

You can also run tests directly:

```bash
go test ./...
```

## Project Layout

- `cmd/relay`: CLI entrypoint and wiring
- `energy`: Aura-Mon HTTP client and CSV parsing
- `database`: InfluxDB 2.x and 3.x writers
- `storage`: file-based timestamp state storage
- `config`: YAML config models and parser
- `release/systemd`: systemd service unit

