package metardata

import (
	"fmt"
	"image/color"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/finack/twinkle/internal/config"
	"github.com/finack/twinkle/internal/display"

	"github.com/gocarina/gocsv"
	"github.com/rs/zerolog/log"
	"golang.org/x/image/colornames"
)

var httpClient = &http.Client{}

// https://www.aviationweather.gov/dataserver/fields?datatype=metar
type Metar struct {
	RawText                   string `csv:"raw_text"`              // The raw METAR
	StationID                 string `csv:"station_id"`            // Station identifier; Always a four character alphanumeric( A-Z, 0-9)
	ObservationTime           string `csv:"observation_time"`      // Time( in ISO8601 date/time format) this METAR was observed.
	Latitude                  string `csv:"latitude"`              // The latitude (in decimal degrees )of the station that reported this METAR
	Longitude                 string `csv:"longitude"`             // The longitude (in decimal degrees) of the station that reported this METAR
	TempC                     string `csv:"temp_c"`                // Air temperature
	DewpointC                 string `csv:"dewpoint_c"`            // Dewpoint temperature
	WindDirDegrees            string `csv:"wind_dir_degrees"`      // Direction from which the wind is blowing.  0 degrees=variable wind direction.
	WindSpeedKt               string `csv:"wind_speed_kt"`         // Wind speed; 0 degree wdir and 0 wspd = calm winds
	WindGustKt                string `csv:"wind_gust_kt"`          // Wind gust
	VisibilityStatuteMi       string `csv:"visibility_statute_mi"` // Horizontal visibility
	AltimInHg                 string `csv:"altim_in_hg"`           // Altimeter
	SeaLevelPressureMb        string `csv:"sea_level_pressure_mb"` // Sea-level pressure
	Corrected                 string `csv:"corrected"`
	Auto                      string `csv:"auto"`
	AutoStation               string `csv:"auto_station"`
	MaintenanceIndicatorOn    string `csv:"maintenance_indicator_on"`
	NoSignal                  string `csv:"no_signal"`
	LightningSensorOff        string `csv:"lightning_sensor_off"`
	FreezingRainSensorOff     string `csv:"freezing_rain_sensor_off"`
	PresentWeatherSensorOff   string `csv:"present_weather_sensor_off"`
	WxString                  string `csv:"wx_string"` // wx_string descriptions
	SkyCover                  string `csv:"sky_cover"`
	CloudBaseftAGL            string `csv:"cloud_base_ft_agl"`
	SkyCover2                 string `csv:"sky_cover"`
	CloudBaseftAGL2           string `csv:"cloud_base_ft_agl"`
	SkyCover3                 string `csv:"sky_cover"`
	CloudBaseftAGL3           string `csv:"cloud_base_ft_agl"`
	SkyCover4                 string `csv:"sky_cover"`
	CloudBaseftAGL4           string `csv:"cloud_base_ft_agl"`
	FlightCategory            string `csv:"flight_category"`               // Flight category of this METAR. Values: VFR|MVFR|IFR|LIFR See http://www.aviationweather.gov/metar/help?page=plot#fltcat"
	ThreeHrPressureTendencyMb string `csv:"three_hr_pressure_tendency_mb"` // Pressure change in the past 3 hours
	MaxTC                     string `csv:"maxT_c"`                        // Maximum air temperature from the past 6 hours
	MinTC                     string `csv:"minT_c"`                        // Minimum air temperature from the past 6 hours
	MaxT24hrC                 string `csv:"maxT24hr_c"`                    // Maximum air temperature from the past 24 hours
	MinT24hrC                 string `csv:"minT24hr_c"`                    // Minimum air temperature from the past 24 hours
	PrecipIn                  string `csv:"precip_in"`                     // Liquid precipitation since the last regular METAR
	Pcp3hrIn                  string `csv:"pcp3hr_in"`                     // Liquid precipitation from the past 3 hours. 0.0005 in = trace precipitation
	Pcp6hrIn                  string `csv:"pcp6hr_in"`                     // Liquid precipitation from the past 6 hours. 0.0005 in = trace precipitation
	Pcp24hrIn                 string `csv:"pcp24hr_in"`                    // Liquid precipitation from the past 24 hours. 0.0005 in = trace precipitation
	SnowIn                    string `csv:"snow_in"`                       // Snow depth on the ground
	VertVisFt                 string `csv:"vert_vis_ft"`                   // Vertical Visibility
	MetarType                 string `csv:"metar_type"`                    // METAR or SPECI
	ElevationM                string `csv:"elevation_m"`                   // The elevation of the station that reported this METAR
}

func FetchRoutine(c config.Config, leds chan display.Pixel) chan bool {
	done := make(chan bool)

	go func() {
		metarRefresh := time.NewTicker(time.Duration(c.MetarRefreshRateS) * time.Second)
		defer metarRefresh.Stop()

		doFetchRoutine(c, leds)
		for {
			select {
			case <-done:
				return
			case <-metarRefresh.C:
				doFetchRoutine(c, leds)
			}
		}
	}()

	return done
}

func flightCategoryToColor(category string) color.RGBA {
	switch strings.ToUpper(category) {
	case "VFR":
		return colornames.Limegreen
	case "MVFR":
		return colornames.Blue
	case "IFR":
		return colornames.Red
	case "LIFR":
		return colornames.Mediumvioletred
	case "", "NULL":
		return colornames.Grey
	default:
		log.Warn().Str("flightCategory", category).Msg("Unknown flightCategory")
		return colornames.Antiquewhite
	}
}

// windyColorFor returns the "windy" hue-shifted variant of a flight category color.
func windyColorFor(category string) color.RGBA {
	switch strings.ToUpper(category) {
	case "VFR":
		return colornames.Yellowgreen
	case "MVFR":
		return colornames.Steelblue
	case "IFR":
		return colornames.Orangered
	case "LIFR":
		return colornames.Deeppink
	default:
		return colornames.Grey
	}
}

// blendColors linearly interpolates between a and b; t is clamped to [0, 1].
func blendColors(a, b color.RGBA, t float64) color.RGBA {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	lerp := func(x, y uint8) uint8 {
		return uint8(float64(x) + t*(float64(y)-float64(x)))
	}
	return color.RGBA{
		R: lerp(a.R, b.R),
		G: lerp(a.G, b.G),
		B: lerp(a.B, b.B),
		A: 0xff,
	}
}

// windAdjustedColor shifts the base color toward the windy variant as effectiveWindKt
// rises from lowKt to highKt, then fades toward white beyond highKt (capped at 40%).
func windAdjustedColor(base, windy color.RGBA, effectiveWindKt, lowKt, highKt float64) color.RGBA {
	if effectiveWindKt < lowKt {
		return base
	}
	if effectiveWindKt < highKt {
		return blendColors(base, windy, (effectiveWindKt-lowKt)/(highKt-lowKt))
	}
	whiteFraction := math.Min((effectiveWindKt-highKt)/15.0, 0.4)
	return blendColors(windy, color.RGBA{R: 255, G: 255, B: 255, A: 255}, whiteFraction)
}

func doFetchRoutine(c config.Config, leds chan display.Pixel) {
	metars, err := getMetars(c.Leds)
	if err != nil {
		log.Error().Err(err).Msg("Could not fetch metars, skipping")
		return
	}

	log.Info().Int("count", len(*metars)).Msg("Fetched Metars")

	for _, metar := range *metars {
		ledNum, ok := c.Stations[metar.StationID]
		if !ok {
			log.Warn().Str("stationID", metar.StationID).Msg("Results included station not found in config")
			continue
		}

		windKt, _ := strconv.ParseFloat(metar.WindSpeedKt, 64)
		gustKt, _ := strconv.ParseFloat(metar.WindGustKt, 64)
		effectiveKt := math.Max(windKt, gustKt)

		log.Debug().
			Str("station", metar.StationID).
			Float64("windKt", windKt).
			Float64("gustKt", gustKt).
			Msg("Wind")

		base := flightCategoryToColor(metar.FlightCategory)
		windy := windyColorFor(metar.FlightCategory)
		col := windAdjustedColor(base, windy, effectiveKt, c.WindLowKt, c.WindHighKt)
		leds <- display.Pixel{Num: ledNum, Color: col}
	}
}

func parseMetarCSV(data []byte) (*[]Metar, error) {
	s := string(data)
	idx := strings.Index(s, "raw_text")
	if idx == -1 {
		return nil, fmt.Errorf("no CSV header found in METAR response")
	}

	stations := []Metar{}
	if err := gocsv.UnmarshalString(s[idx:], &stations); err != nil {
		log.Error().Err(err).Msg("Could not unmarshal CSV")
		log.Error().Str("body", s[idx:]).Msg("CSV Data")
		return nil, err
	}
	return &stations, nil
}

func getMetars(s map[int]string) (*[]Metar, error) {
	data, err := fetchMetars(s)
	if err != nil {
		return nil, err
	}
	return parseMetarCSV(data)
}

var metarBaseURL = "https://www.aviationweather.gov/api/data/dataserver"

func fetchMetars(s map[int]string) ([]byte, error) {
	var stations []string
	for _, station := range s {
		stations = append(stations, station)
	}

	url := metarBaseURL + "?dataSource=metars&requestType=retrieve&format=csv"
	url += "&mostRecentForEachStation=true&hoursBeforeNow=4"
	url += "&stationString="
	url += strings.Join(stations, ",")

	resp, err := httpClient.Get(url)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("Unable to fetch Metar")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("HTTP expected %v got %v", http.StatusOK, resp.StatusCode)
		log.
			Error().
			Err(err).
			Int("httpstatus", resp.StatusCode).
			Int("expectedHttpStatus", http.StatusOK).
			Msg("Received an unexpected HTTP Status")
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Unable to parse HTTP Body")
		return nil, err
	}

	return data, nil
}
