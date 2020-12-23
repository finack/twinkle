
[![Go Report Card](https://goreportcard.com/badge/github.com/finack/twinkle)](https://goreportcard.com/report/github.com/finack/twinkle)
[![license](https://img.shields.io/github/license/finack/twinkle.svg)](https://github.com/finack/twinkle)

# Twinkle - Metar Map

Inspired by [Philip's Metar Map](https://slingtsi.rueker.com/making-a-led-powered-metar-map-for-your-wall/), this project controls a handfull of LEDs mounted behind a sectional chart.

[TODO List](TODO.md)

## Installation

Twinkle is designed to run on a raspberry pi that has the [rpi_ws281x library](https://github.com/jgarff/rpi_ws281x) installed. This assume you are running on an ubuntu distro.

### rpi_ws281x library installation

From the raspberry

``` shell
sudo apt-get install cmake

git clone https://github.com/jgarff/rpi_ws281x

cd rpi_ws281x
mkdir build
cd build

cmake -D BUILD_SHARED=OFF -D BUILD_TEST=ON ..

sudo make install
```

You should now have `/usr/local/lib/libws2811.a` and headers in `/usr/local/include/ws2811`. If not, check out the [build documentation](https://github.com/jgarff/rpi_ws281x#build).

### Twinkle

Clone this repo and run `make build` to see if you go toolchain is hapoy. You might need to install go and more tools.

Then tweak `config.yaml`

Also check out some other commands:
* **`make run`** : fire up Twinkle from source
* **`make setup`** : Very opinionated and fragile way to run twinkle via `systemd`.
* **`make [enable|disable]`** : Tell `systemd` to run twinkle on startup (or not); needs setup to run first
* **`make [start|stop|status]`** : Find out how `systemd` feels about twinkle, start twinkle or stop it

