package metardata

import (
	"errors"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/finack/twinkle/internal/config"
	"github.com/finack/twinkle/internal/display"
	"github.com/gocarina/gocsv"
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
	windSpeedKt               int     `csv:"wind_speed_kt"`                 // Wind speed; 0 degree wdir and 0 wspd = calm winds
	windGustKt                int     `csv:"wind_gust_kt"`                  // Wind gust
	VisibilityStatuteMi       float32 `csv:"visibility_statute_mi"`         // Horizontal visibility
	AltimInHg                 float32 `csv:"altim_in_hg"`                   // Altimeter
	SeaLevelPressureMb        float32 `csv:"sea_level_pressure_mb"`         // Sea-level pressure
	QualityControlFlags       string  `csv:"quality_control_flags"`         // Quality control flags (see below) provide useful information about the METAR station(s) that provide the data.
	WxString                  string  `csv:"wx_string"`                     // wx_string descriptions
	SkyCondition              string  `csv:"sky_condition"`                 // sky_cover - up to four levels of sky cover and base can be reported under the sky_conditions field; OVX present when vert_vis_ft is reported.  Allowed values: SKC|CLR|CAVOK|FEW|SCT|BKN|OVC|OVX"
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
	log.Println("Fetching metars")
	metars, err := getMetars(c.Leds)
	if err != nil {
		log.Printf("MAIN:Could not fetch metars: %v", err)
	}

	for _, metar := range *metars {
		var ledNum int = -1
		ledNum = c.Stations[metar.StationID]
		if ledNum < 0 {
			log.Printf("Could not find station %v", metar.StationID)
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
			log.Printf("Unknown flightCategory %v for %v", metar.FlightCategory, metar)
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
		log.Fatalf("Could not unmarshal CSV: %v", err)
	}

	return &stations, nil
}

func fetchMetars(s map[int]string) ([]byte, error) {

	stations := make([]string, 0)
	for _, station := range s {
		stations = append(stations, station)
	}

	url := "https://www.aviationweather.gov/adds/dataserver_current/httpparam?dataSource=metars&requestType=retrieve&format=csv"
	url += "&mostRecentForEachStation=true&hoursBeforeNow=4"
	url += "&stationString="
	url += strings.Join(stations, ",")

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Could not fetch %v : %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Sprintf("[fetchMetar][ERROR] HTTP expected %v got %v", http.StatusOK, resp.StatusCode)
		log.Printf(err)
		return nil, errors.New(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[fetchMetar][ERROR] Read body: %v", err)
		return nil, err
	}

	return data, nil
}
