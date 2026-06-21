package display

import (
	"errors"
	"image/color"
	"time"

	"github.com/finack/twinkle/internal/config"
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

func newWithEngine(ws wsEngine) *Leds {
	return &Leds{Ws: ws}
}

func UpdateRoutine(c config.Config) (chan bool, chan Pixel) {
	done := make(chan bool)
	ledChannel := make(chan Pixel)

	loc, err := time.LoadLocation(c.Locale)
	if err != nil {
		log.Fatal().Err(err).Str("locale", c.Locale).Caller().Msg("Could not load timezone")
	}

	// Extract only the scalars needed so the goroutine closure doesn't retain
	// the Leds/Stations maps for the process lifetime.
	ledCount := c.LedCount
	ledRefreshRateMS := c.LedRefreshRateMS
	brightness := c.Brightness
	nightBrightness := c.NightBrightness
	latitude := c.Latitude
	longitude := c.Longitude

	go func() {
		display := make([]color.RGBA, ledCount)
		var buffer []Pixel

		leds, err := New(brightness, ledCount)
		if err != nil {
			log.Fatal().Err(err).Caller().Msg("Could not start connection to LEDS")
		}

		ledRefreshRate := time.NewTicker(time.Duration(ledRefreshRateMS) * time.Millisecond)
		defer ledRefreshRate.Stop()

		brightnessRefresh := time.NewTicker(10 * time.Second)
		defer brightnessRefresh.Stop()

		var (
			riseSetDate time.Time
			cachedRise  time.Time
			cachedSet   time.Time
		)

		for {
			select {
			case <-done:
				leds.Clear()
				leds.Ws.Fini()
				return
			case m := <-ledChannel:
				if display[m.Num] != m.Color {
					buffer = append(buffer, m)
				}
			case <-brightnessRefresh.C:
				now := time.Now().In(loc)
				today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
				if !today.Equal(riseSetDate) {
					cachedRise, cachedSet, err = calcRiseSet(now, longitude, latitude)
					if err != nil {
						continue
					}
					riseSetDate = today
				}
				b := calcBrightness(now, cachedRise, cachedSet, brightness, nightBrightness)
				leds.Ws.SetBrightness(0, b)
				if err := leds.Ws.Render(); err != nil {
					log.Error().Err(err).Caller().Msg("Issue rendering brightness change")
				}
				log.Debug().Int("brightness", b).Msg("Updated brightness")
			case <-ledRefreshRate.C:
				if len(buffer) == 0 {
					continue
				}

				log.Debug().Int("ledCount", len(buffer)).Msg("Updating Display")
				for _, p := range buffer {
					leds.Display(p.Num, p.Color)
					display[p.Num] = p.Color
				}
				if err := leds.Ws.Render(); err != nil {
					log.Error().Err(err).Caller().Msg("Issue rendering to LEDS")
				}
				buffer = nil
			}
		}
	}()

	return done, ledChannel
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

func (l *Leds) Display(num int, c color.RGBA) {
	l.Ws.Leds(0)[num] = ParseRGBAtoUint32(c)
}

func ParseRGBAtoUint32(c color.RGBA) uint32 {
	return uint32(c.R)<<16 | uint32(c.G)<<8 | uint32(c.B)
}

// From https://stackoverflow.com/questions/54197913/parse-hex-string-to-image-color
func ParseHexColor(s string) (c color.RGBA, err error) {
	errInvalidFormat := errors.New("image is not a valid format")
	c.A = 0xff

	if len(s) == 0 || s[0] != '#' {
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
