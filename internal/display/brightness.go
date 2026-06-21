package display

import (
	"errors"
	"time"

	sunrise "github.com/nathan-osman/go-sunrise"
	"github.com/rs/zerolog/log"
)

func calcRiseSet(now time.Time, long, lat float64) (rise time.Time, set time.Time, err error) {
	rise, set = sunrise.SunriseSunset(lat, long, now.Year(), now.Month(), now.Day())

	empty := time.Time{}
	if rise == empty || set == empty {
		log.Error().Time("sunrise", rise).Time("sunset", set).Caller().Msg("Sunrise/Sunset not calculated")
		err = errors.New("Sunrise/Sunset not calculated")
		return
	}

	rise = rise.In(now.Location())
	set = set.In(now.Location())

	log.Debug().Time("sunrise", rise).Time("sunset", set).Msg("Sunrise & Sunset info")
	return
}

// calcBrightness interpolates between dayBrightness and nightBrightness over a 30-minute
// transition window centered on each horizon crossing.
func calcBrightness(now, rise, set time.Time, dayBrightness, nightBrightness int) int {
	const window = 30 * time.Minute

	riseStart := rise.Add(-window)
	riseEnd := rise.Add(window)
	setStart := set.Add(-window)
	setEnd := set.Add(window)

	switch {
	case now.Before(riseStart):
		return nightBrightness
	case now.Before(riseEnd):
		fraction := float64(now.Sub(riseStart)) / float64(2*window)
		return nightBrightness + int(float64(dayBrightness-nightBrightness)*fraction)
	case now.Before(setStart):
		return dayBrightness
	case now.Before(setEnd):
		fraction := float64(now.Sub(setStart)) / float64(2*window)
		return dayBrightness - int(float64(dayBrightness-nightBrightness)*fraction)
	default:
		return nightBrightness
	}
}
