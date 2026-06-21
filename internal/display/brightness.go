package display

import (
	"errors"
	"time"

	sunrise "github.com/nathan-osman/go-sunrise"
	"github.com/rs/zerolog/log"
)

func calcRiseSet(long, lat float64, locale string) (t time.Time, rise time.Time, set time.Time, err error) {
	location, err := time.LoadLocation(locale)
	if err != nil {
		log.Error().Err(err).Caller().Msg("Could not lookup location")
		return
	}

	t = time.Now().In(location)
	rise, set = sunrise.SunriseSunset(lat, long, t.Year(), t.Month(), t.Day())

	empty := time.Time{}
	if rise == empty || set == empty {
		log.Error().Time("sunrise", rise).Time("sunset", set).Caller().Msg("Sunrise/Sunset not calculated")
		err = errors.New("Sunrise/Sunset not calculated")
		return
	}

	rise = rise.In(location)
	set = set.In(location)

	log.Debug().Time("sunrise", rise).Time("sunset", set).Msg("Sunrise & Sunset info")
	return
}

// calcBrightness interpolates between dayBrightness and nightBrightness over a 30-minute
// transition window centered on each horizon crossing.
func calcBrightness(now, rise, set time.Time, dayBrightness, nightBrightness int) int {
	const window = 30 * time.Minute

	switch {
	case now.Before(rise.Add(-window)):
		return nightBrightness
	case now.Before(rise.Add(window)):
		elapsed := now.Sub(rise.Add(-window))
		fraction := float64(elapsed) / float64(2*window)
		return nightBrightness + int(float64(dayBrightness-nightBrightness)*fraction)
	case now.Before(set.Add(-window)):
		return dayBrightness
	case now.Before(set.Add(window)):
		elapsed := now.Sub(set.Add(-window))
		fraction := float64(elapsed) / float64(2*window)
		return dayBrightness - int(float64(dayBrightness-nightBrightness)*fraction)
	default:
		return nightBrightness
	}
}
