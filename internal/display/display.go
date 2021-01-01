package display

import (
	"errors"
	"image/color"
	"time"

	"github.com/finack/twinkle/internal/config"
	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
  "github.com/rs/zerolog/log"
)

type wsEngine interface {
	Init() error
	Render() error
	Wait() error
	Fini()
	Leds(channel int) []uint32
	SetBrightness(channel int, brightness int)
}

type Leds struct {
	Ws wsEngine
}

type Pixel struct {
	Num   int
	Color color.RGBA
}

func New(brightness int, ledcount int) (*Leds, error) {
	opt := ws2811.DefaultOptions
	opt.Channels[0].Brightness = brightness
	opt.Channels[0].LedCount = ledcount
	opt.Channels[0].StripeType = ws2811.WS2811StripRGB

	dev, err := ws2811.MakeWS2811(&opt)
	if err != nil {
		log.
      Fatal().
      Err(err).
      Caller().
      Msg("Could not configure LEDS")
	}

	err = dev.Init()
	if err != nil {
		log.
      Fatal().
      Err(err).
      Caller().
		  Msg("Could not init LEDS")
	}

	leds := &Leds{Ws: dev}

	leds.Clear()

	return leds, nil
}

func UpdateRoutine(c config.Config) (chan bool, chan bool, chan Pixel) {
	done := make(chan bool)
	refresh := make(chan bool)

	ledChannel := make(chan Pixel)

	go func() {
		display := make([]color.RGBA, c.LedCount)
		var buffer []Pixel

		leds, err := New(c.Brightness, c.LedCount)
		if err != nil {
			log.
        Fatal().
        Err(err).
        Caller().
        Msg("Could not start connection to LEDS")
		}

		ledRefreshRate := time.NewTicker(time.Duration(c.LedRefreshRateMS) * time.Millisecond)
		defer ledRefreshRate.Stop()

		for {
			select {
			case <-done:
				leds.Clear()
				leds.Ws.Fini()
				return
			case m := <-ledChannel:
				if display[m.Num] != m.Color {
					buffer = append(buffer, m)
					// log.Printf("setting led[%v][%#v]", m.Num, m.Color)
				}
			case <-refresh:
				for n, c := range display {
					err := leds.Display(n, c)
					if err != nil {
						log.Error().Caller().Msgf("Issue setting pix %v : %v", n, err)
            continue
					}

					err = leds.Ws.Render()
					if err != nil {
						log.Error().Err(err).Caller().Msg("Issue rendering to LEDS")
            continue
					}
				}

			case <-ledRefreshRate.C:
        if len(buffer) <= 0 {
          continue
        }

        log.Debug().Int("ledCount", len(buffer)).Msg("Updating Display")
				for _, p := range buffer {
					err := leds.Display(p.Num, p.Color)
					if err != nil {
						log.Error().Err(err).Caller().Int("pixel", p.Num).Msg("Issue setting pixel")
            continue
					}

					display[p.Num] = p.Color

					err = leds.Ws.Render()
					if err != nil {
						log.Error().Err(err).Caller().Msg("Issue rendering to LEDS")
					}
				}

				buffer = nil
			}
		}
	}()

	return done, refresh, ledChannel
}

func (l *Leds) Clear() error {
	for i := 0; i < len(l.Ws.Leds(0)); i++ {
		l.Ws.Leds(0)[i] = 0
		if err := l.Ws.Render(); err != nil {
			return err
		}
	}
	return nil
}

func (l *Leds) Display(num int, c color.RGBA) error {

	l.Ws.Leds(0)[num] = ParseRGBAtoUint32(c)

	if err := l.Ws.Render(); err != nil {
		return err
	}

	return nil
}

func ParseRGBAtoUint32(c color.RGBA) uint32 {
	return (((uint32(c.R) & 0x0FF) << 16) | ((uint32(c.G) & 0x0ff) << 8) | (uint32(c.B) & 0x0ff))
}

// From https://stackoverflow.com/questions/54197913/parse-hex-string-to-image-color
func ParseHexColor(s string) (c color.RGBA, err error) {
	errInvalidFormat := errors.New("Image is not a valid format")
	c.A = 0xff

	if s[0] != '#' {
		return c, errInvalidFormat
	}

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		}
		err = errInvalidFormat
		return 0
	}

	switch len(s) {
	case 7:
		c.R = hexToByte(s[1])<<4 + hexToByte(s[2])
		c.G = hexToByte(s[3])<<4 + hexToByte(s[4])
		c.B = hexToByte(s[5])<<4 + hexToByte(s[6])
	case 4:
		c.R = hexToByte(s[1]) * 17
		c.G = hexToByte(s[2]) * 17
		c.B = hexToByte(s[3]) * 17
	default:
		err = errInvalidFormat
	}
	return
}
