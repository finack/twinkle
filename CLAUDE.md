# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this project does

Twinkle controls WS2811/WS2812 RGB LED strips mounted behind a sectional aviation chart on a Raspberry Pi. Each LED represents an airport; its color reflects the current METAR flight category (VFR=lime green, MVFR=blue, IFR=red, LIFR=medium violet red, no data=grey).

## Commands

```bash
# Run tests (works on macOS; display package compiles via stub, not hardware)
go test ./internal/...

# Run a single test
go test ./internal/metardata/ -run TestParseMetarCSV

# Run tests with verbose output
go test -v ./internal/...

# Vet
go vet ./internal/...

# Build for local platform (only useful on Linux/Pi with rpi_ws281x installed)
make build

# Cross-compile for Raspberry Pi (requires Docker buildx)
make docker-setup   # one-time setup
make docker-build   # produces build/twinkle-armarm32v6

# Deploy and run via systemd on the Pi
make setup          # copies binary + config + service file to /home/twinkle
make enable         # configure systemd autostart
make start / stop / status
```

## Architecture

Three goroutines communicate over channels, started in `cmd/server/main.go`:

```
FetchRoutine  →  ledChannel (chan Pixel)  →  UpdateRoutine  →  WS2811 hardware
                                                 ↑
AutomaticDimmer  →  setBrightness (chan int) ─────┘
```

- **`internal/metardata`** — fetches METAR CSV from `aviationweather.gov` every N seconds, maps `FlightCategory` to a color, and sends `display.Pixel{Num, Color}` values onto `ledChannel`.
- **`internal/display`** — buffers incoming `Pixel` updates and renders them to the LED strip on a timer tick. `UpdateRoutine` owns the hardware lifecycle.
- **`internal/config`** — parses `config.yaml`; builds both the `Leds map[int]string` (LED index → station) and the reverse `Stations map[string]int` (station → LED index).
- **`internal/display/brightness.go`** — `AutomaticDimmer` adjusts brightness via `setBrightness` channel based on sunrise/sunset.
- **`internal/signals`** — catches SIGINT/SIGTERM and sends stop signals to all goroutines for a clean shutdown.

## Build constraints and hardware dependency

`rpi-ws281x-go` is a cgo library that requires the native `rpi_ws281x` C library (only available on the Pi). The display package uses a build-tag split:

- `display_hardware.go` — `//go:build linux` — real `New()` using ws2811
- `display_stub.go` — `//go:build !linux` — stub `New()` that returns an error

This lets the package compile and be tested on macOS. The Docker build in `docker/app-builder/Dockerfile` compiles the C library and cross-compiles the Go binary for ARM.

## Known bug

`brightness.go:68` reads `set = rise.In(location)` but should be `set = set.In(location)`. This makes sunset always equal sunrise, so `AutomaticDimmer` never correctly detects after-sunset. `TestCalcRiseSet_ValidLocale` in `internal/display/brightness_test.go` intentionally fails to document this.

## Testing approach

- Tests use inline CSV strings (not `sample.csv`, which may be stale).
- `internal/display` tests inject a `mockWsEngine` via the unexported `newWithEngine()` constructor.
- `internal/metardata` tests inject an `httptest.Server` via the package-level `httpClient` and `metarBaseURL` vars.
- All test files live in the same package as the code they test (not `_test` suffix packages), so unexported functions are accessible directly.
