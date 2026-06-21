//go:build linux

package display

import (
	"fmt"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

func New(brightness int, ledcount int) (*Leds, error) {
	opt := ws2811.DefaultOptions
	opt.Channels[0].Brightness = brightness
	opt.Channels[0].LedCount = ledcount
	opt.Channels[0].StripeType = ws2811.WS2811StripRGB
	opt.Channels[0].GpioPin = 12

	dev, err := ws2811.MakeWS2811(&opt)
	if err != nil {
		return nil, fmt.Errorf("configure LEDs: %w", err)
	}

	if err = dev.Init(); err != nil {
		return nil, fmt.Errorf("init LEDs: %w", err)
	}

	leds := newWithEngine(dev)
	leds.Clear()
	return leds, nil
}
