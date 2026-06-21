//go:build linux

package display

import (
	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
	"github.com/rs/zerolog/log"
)

func New(brightness int, ledcount int) (*Leds, error) {
	opt := ws2811.DefaultOptions
	opt.Channels[0].Brightness = brightness
	opt.Channels[0].LedCount = ledcount
	opt.Channels[0].StripeType = ws2811.WS2811StripRGB
	opt.Channels[0].GpioPin = 12

	dev, err := ws2811.MakeWS2811(&opt)
	if err != nil {
		log.Fatal().Err(err).Caller().Msg("Could not configure LEDS")
	}

	err = dev.Init()
	if err != nil {
		log.Fatal().Err(err).Caller().Msg("Could not init LEDS")
	}

	leds := newWithEngine(dev)
	leds.Clear()

	return leds, nil
}
