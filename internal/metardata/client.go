package metardata

import (
	"errors"
	"fmt"
	"image/color"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/finack/twinkle/internal/config"
	"github.com/finack/twinkle/internal/display"

	"github.com/gocarina/gocsv"
	"github.com/rs/zerolog/log"
	"golang.org/x/image/colornames"
)

// https://www.aviationweather.gov/dataserver/fields?datatype=metar
type Metar struct {
	RawText                   string  `csv:"raw_text"`                      // The raw METAR
	StationID                 string  `csv:"station_id"`                    // Station identifier; Always a four character alphanumeric( A-Z, 0-9)
	ObservationTime           string  `csv:"observation_time"`              // Time( in ISO8601 date/time format) this METAR was observed.
	Latitude                  float32 `csv:"latitude"`                      // The latitude (in decimal degrees )of the station that reported this METAR
	Longitude                 float32 `csv:"longitude"`                     // The longitude (in decimal degrees) of the station that reported this METAR
	TempC                     float32 `csv:"temp_c"`                        // Air temperature
	DewpointC                 float32 `csv:"dewpoint_c"`                    // Dewpoint temperature
	WindDirDegrees            int     `csv:"wind_dir_degrees"`              // Direction from which the wind is blowing.  0 degrees=variable wind direction.
	windSpeedKt               string  `csv:"wind_speed_kt"`                 // Wind speed; 0 degree wdir and 0 wspd = calm winds
	windGustKt                string  `csv:"wind_gust_kt"`                  // Wind gust
	VisibilityStatuteMi       string  `csv:"visibility_statute_mi"`         // Horizontal visibility
	AltimInHg                 string  `csv:"altim_in_hg"`                   // Altimeter
	SeaLevelPressureMb        string  `csv:"sea_level_pressure_mb"`         // Sea-level pressure
	Corrected                 string  `csv:"corrected"`
	Auto                      string  `csv:"auto"`
	AutoStation               string  `csv:"auto_station"`
	MaintenanceIndicatorOn    string  `csv:"maintenance_indicator_on"`
	NoSignal                  string  `csv:"no_signal"`
	LightningSensorOff        string  `csv:"lightning_sensor_off"`
	FreezingRainSensorOff     string  `csv:"freezing_rain_sensor_off"`
	PresentWeatherSensorOff   string  `csv:"present_weather_sensor_off"`
	WxString                  string  `csv:"wx_string"`                     // wx_string descriptions
	SkyCover                  string  `csv:"sky_cover"`
	CloudBaseftAGL            float32 `csv:"cloud_base_ft_agl"`
	SkyCover2                 string  `csv:"sky_cover"`
	CloudBaseftAGL2           float32 `csv:"cloud_base_ft_agl"`
	SkyCover3                 string  `csv:"sky_cover"`
	CloudBaseftAGL3           float32 `csv:"cloud_base_ft_agl"`
	SkyCover4                 string  `csv:"sky_cover"`
	CloudBaseftAGL4           float32 `csv:"cloud_base_ft_agl"`
	FlightCategory            string  `csv:"flight_category"`               // Flight category of this METAR. Values: VFR|MVFR|IFR|LIFR See http://www.aviationweather.gov/metar/help?page=plot#fltcat"
	ThreeHrPressureTendencyMb float32 `csv:"three_hr_pressure_tendency_mb"` // Pressure change in the past 3 hours
	MaxTC                     float32 `csv:"maxT_c"`                        // Maximum air temperature from the past 6 hours
	MinTC                     float32 `csv:"minT_c"`                        // Minimum air temperature from the past 6 hours
	MaxT24hrC                 float32 `csv:"maxT24hr_c"`                    // Maximum air temperature from the past 24 hours
	MinT24hrC                 float32 `csv:"minT24hr_c"`                    // Minimum air temperature from the past 24 hours
	PrecipIn                  float32 `csv:"precip_in"`                     // Liquid precipitation since the last regular METAR
	Pcp3hrIn                  float32 `csv:"pcp3hr_in"`                     // Liquid precipitation from the past 3 hours. 0.0005 in = trace precipitation
	Pcp6hrIn                  float32 `csv:"pcp6hr_in"`                     // Liquid precipitation from the past 6 hours. 0.0005 in = trace precipitation
	Pcp24hrIn                 float32 `csv:"pcp24hr_in"`                    // Liquid precipitation from the past 24 hours. 0.0005 in = trace precipitation
	SnowIn                    float32 `csv:"snow_in"`                       // Snow depth on the ground
	VertVisFt                 int     `csv:"vert_vis_ft"`                   // Vertical Visibility
	MetarType                 string  `csv:"metar_type"`                    // METAR or SPECI
	ElevationM                int     `csv:"elevation_m"`                   // The elevation of the station that reported this METAR
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

func doFetchRoutine(c config.Config, leds chan display.Pixel) {

	metars, err := getMetars(c.Leds)
	if err != nil {
		log.Error().Err(err).Msg("Could not fetch metars, skipping")
    return
	}

	log.Info().Int("count", len(*metars)).Msg("Fetched Metars")

	for _, metar := range *metars {
		var ledNum int = -1
		ledNum = c.Stations[metar.StationID]
		if ledNum < 0 {
			log.Warn().Str("stationID", metar.StationID).Msg("Results included station not found in config")
			continue
		}

		var color color.RGBA
		flightCategory := strings.ToUpper(metar.FlightCategory)
		switch flightCategory {
		case "VFR":
			color = colornames.Limegreen
		case "MVFR":
			color = colornames.Blue
		case "IFR":
			color = colornames.Red
		case "LIFR":
			color = colornames.Mediumvioletred
		case "":
			color = colornames.Grey
		default:
			log.Warn().Str("flightCategory", metar.FlightCategory).Msg("Unknown flightCategory")
			color = colornames.Antiquewhite
		}
		leds <- display.Pixel{Num: ledNum, Color: color}
	}
	return
}

func getMetars(s map[int]string) (*[]Metar, error) {

	data, err := fetchMetars(s)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	var csvData strings.Builder

	foundHeader := false
	for _, line := range lines {
		if strings.HasPrefix(line, "raw_text") == true {
			foundHeader = true
		}

		if foundHeader == true {
			line += "\n"
			csvData.WriteString(line)
		} else {
			continue
		}
	}

	stations := []Metar{}

	err = gocsv.UnmarshalString(csvData.String(), &stations)
	if err != nil {
		log.Error().Err(err).Msg("Could not unmarshal CSV")
		log.Error().Str("body", csvData.String()).Msg("CSV Data")
    return nil, err
	}

	return &stations, nil
}

func fetchMetars(s map[int]string) ([]byte, error) {

	stations := make([]string, 0)
	for _, station := range s {
		stations = append(stations, station)
	}

	url := "https://www.aviationweather.gov/api/data/dataserver?dataSource=metars&requestType=retrieve&format=csv"
	url += "&mostRecentForEachStation=true&hoursBeforeNow=4"
	url += "&stationString="
	url += strings.Join(stations, ",")

	resp, err := http.Get(url)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("Unable to fetch Metar")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := errors.New(fmt.Sprintf("HTTP expected %v got %v", http.StatusOK, resp.StatusCode))
		log.
			Error().
			Err(err).
			Int("httpstatus", resp.StatusCode).
			Int("expectedHttpStatus", http.StatusOK).
			Msg("Received an unexpected HTTP Status")
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Unable to parse HTTP Body")
		return nil, err
	}


	return data, nil
}
