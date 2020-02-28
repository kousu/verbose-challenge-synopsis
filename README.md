# TouchTunes Batch Device Updater

There are thousands of TouchTunes players deployed as consumer-premises equipment.
Their software consists of apps which run and can be updated semi-independently.
This tool batch updates the software on these devices.

## Usage

### Synopsis

```
update-player-firmwares [-h] [-v] [application_id=vM.m.p ...] device_csv
```

To run updates, first collect the apps to update and their desired new versions.
The versions should be in [semver](https://semver.org/) format: `vMAJOR.MINOR.PATCH`. You can find the list of available apps with their tagged versions on the [project server](https://gitlab.intra.touchtunes.com).

Second, collect the devices to target. Provide them in a CSV file by MAC address, say "devices.csv":

```{csv}
mac_addresses
8c:50:ff:70:f5:94
9c:76:8c:19:57:bf
40:67:56:ed:e7:dc
```

You will also need to have an authentication token from the
[intranet OAuth server](https://intra.touchtunes.example.com/authorize/updates).
Provide it via an environment variable:

```{sh}
export TOUCHTUNES_AUTH_TOKEN="<your-token>"
```

or on Windows:

```{cmd}
set TOUCHTUNES_AUTH_TOKEN="<your-token>"
```

Run the batch job like this

```{sh}
update-player-firmwares music_app=v1.4.10 diagnostic_app=v1.2.6 settings_app=v1.1.5 devices.csv
```

You may use "-" instead of the filename to read from stdin instead:

```{sh}
update-player-firmwares music_app=v1.4.10 diagnostic_app=v1.2.6 settings_app=v1.1.5 - <<EOF
mac_addresses
8c:50:ff:70:f5:94
9c:76:8c:19:57:bf
40:67:56:ed:e7:dc
EOF
```

Options:

* `-h`: show short usage
* `-v`: be more verbose about the operations in progress

Environment variables:

* `TOUCHTUNES_FLEET_API`: the RESTful endpoint that manages
       the TouchTunes fleet; optional, only to override the default.
* `TOUCHTUNES_AUTH_TOKEN`: the token authorizing usage of
       the fleet API; required.

## Installation

To install from source, first get [go 1.12+](https://golang.org/) installed (e.g. `brew install go`, `apt-get install go`),
making sure `GOBIN` is defined, and that `$GOBIN` is part of your system `PATH`. Then

```{sh}
go install
```

will produce `update-player-firmwares`.

## Development

To run and test the app without installation, use `go run`. It is also wise to point the app at a staging API during development. Hence,

```
export TOUCHTUNES_FLEET_API=http://localhost:6565  # or "set" if on Windows
go run update-player-firmwares.go [options] [apps ...] device_csv
```

Please remember to run `gofmt -w` regularly over your code.

### Tests

To run unit tests, just

```{sh}
go test -v
```

The tests are in `update-player-firmwares_test.go`
