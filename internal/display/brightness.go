package display

import (
	"errors"
	"time"

	"github.com/finack/twinkle/internal/config"

	sunrise "github.com/nathan-osman/go-sunrise"
	"github.com/rs/zerolog/log"
)

func AutomaticDimmer(c config.Config, setBrightness chan int) chan bool {

	done := make(chan bool)

	go func() {
		brightnessRefresh := time.NewTicker(10 * time.Second)

		for {
			select {
			case <-done:
				return
			case <-brightnessRefresh.C:
				now, rise, set, err := calcRiseSet(c.Longitude, c.Latitude, c.Locale)
				if err != nil {
					continue
				}

				if now.After(set) {
					log.Info().Msg("After sunset")
					setBrightness <- (c.Brightness / 2)

				} else if now.Before(rise) {
					log.Info().Msg("Before sunset")
					setBrightness <- (c.Brightness / 2)
				} else {
					setBrightness <- c.Brightness
				}
			}
		}
	}()

	return done
}

func calcRiseSet(long, lat float64, locale string) (t time.Time, rise time.Time, set time.Time, err error) {

	location, err := time.LoadLocation(locale)

	if err != nil {
		log.Error().Err(err).Caller().Msg("Could not lookup location")
		return
	}

	t = time.Now().In(location)

	rise, set = sunrise.SunriseSunset(lat, lat, t.Year(), t.Month(), t.Day())
	empty := time.Time{}

	if (rise == empty) || (set == empty) {
		log.Error().Time("sunrise", rise).Time("sunset", set).Caller().Msg("Sunrise/Sunset not calculated")
		err = errors.New("Sunrise/Sunset not calculated")
		return
	}

	rise = rise.In(location)
	set = rise.In(location)

	log.Debug().Time("sunrise", rise).Time("sunset", set).Msg("Sunrise & Sunset info")
	return
}
