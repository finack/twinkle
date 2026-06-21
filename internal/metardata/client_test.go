package metardata

import (
	"image/color"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/image/colornames"
)

// csvHeader is the real Aviation Weather API CSV header line.
const csvHeader = "raw_text,station_id,observation_time,latitude,longitude,temp_c,dewpoint_c,wind_dir_degrees,wind_speed_kt,wind_gust_kt,visibility_statute_mi,altim_in_hg,sea_level_pressure_mb,corrected,auto,auto_station,maintenance_indicator_on,no_signal,lightning_sensor_off,freezing_rain_sensor_off,present_weather_sensor_off,wx_string,sky_cover,cloud_base_ft_agl,sky_cover,cloud_base_ft_agl,sky_cover,cloud_base_ft_agl,sky_cover,cloud_base_ft_agl,flight_category,three_hr_pressure_tendency_mb,maxT_c,minT_c,maxT24hr_c,minT24hr_c,precip_in,pcp3hr_in,pcp6hr_in,pcp24hr_in,snow_in,vert_vis_ft,metar_type,elevation_m"

// makeRow builds a minimal CSV data row with only station_id and flight_category filled in.
// Header has 44 columns; flight_category is column 31 (29 commas after station_id, 13 after category).
func makeRow(stationID, flightCategory string) string {
	return "RAW TEXT," + stationID + ",,,,,,,,,,,,,,,,,,,,,,,,,,,,," + flightCategory + ",,,,,,,,,,,,,"
}

// preamble simulates the non-CSV lines the Aviation Weather API prepends before the header.
const preamble = "No errors\nNo warnings\n24 results\n\n"

func TestParseMetarCSV(t *testing.T) {
	csv := preamble + csvHeader + "\n" +
		makeRow("KOAK", "VFR") + "\n" +
		makeRow("KSFO", "IFR") + "\n" +
		makeRow("KHAF", "LIFR") + "\n" +
		makeRow("KAPC", "MVFR") + "\n"

	metars, err := parseMetarCSV([]byte(csv))
	if err != nil {
		t.Fatalf("parseMetarCSV error: %v", err)
	}
	if len(*metars) != 4 {
		t.Fatalf("got %d metars, want 4", len(*metars))
	}

	cases := []struct {
		idx      int
		station  string
		category string
	}{
		{0, "KOAK", "VFR"},
		{1, "KSFO", "IFR"},
		{2, "KHAF", "LIFR"},
		{3, "KAPC", "MVFR"},
	}
	for _, tc := range cases {
		m := (*metars)[tc.idx]
		if m.StationID != tc.station {
			t.Errorf("[%d] StationID: got %q, want %q", tc.idx, m.StationID, tc.station)
		}
		if m.FlightCategory != tc.category {
			t.Errorf("[%d] FlightCategory: got %q, want %q", tc.idx, m.FlightCategory, tc.category)
		}
	}
}

func TestParseMetarCSV_MissingHeader(t *testing.T) {
	// No raw_text line — should return empty slice (header never found).
	csv := preamble + "not,a,real,header\nsome,data,row\n"
	metars, err := parseMetarCSV([]byte(csv))
	// Either no error with empty result, or an error — both are acceptable.
	if err != nil {
		return
	}
	if len(*metars) != 0 {
		t.Errorf("expected 0 metars when header is missing, got %d", len(*metars))
	}
}

func TestParseMetarCSV_Empty(t *testing.T) {
	// Empty input has no header line, so parseMetarCSV returns empty or errors — both acceptable.
	metars, err := parseMetarCSV([]byte(""))
	if err == nil && len(*metars) != 0 {
		t.Errorf("expected 0 metars on empty input, got %d", len(*metars))
	}
}

// flightCategoryToColor tests

func TestFlightCategoryToColor(t *testing.T) {
	tests := []struct {
		input string
		want  color.RGBA
	}{
		{"VFR", colornames.Limegreen},
		{"vfr", colornames.Limegreen},
		{"MVFR", colornames.Blue},
		{"IFR", colornames.Red},
		{"LIFR", colornames.Mediumvioletred},
		{"", colornames.Grey},
		{"UNKNOWN", colornames.Antiquewhite},
		{"XYZ", colornames.Antiquewhite},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := flightCategoryToColor(tt.input)
			if got != tt.want {
				t.Errorf("flightCategoryToColor(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

// fetchMetars tests using httptest

func TestFetchMetars_Success(t *testing.T) {
	body := preamble + csvHeader + "\n" + makeRow("KOAK", "VFR") + "\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
	defer srv.Close()

	// Inject test server URL and client
	origURL := metarBaseURL
	origClient := httpClient
	metarBaseURL = srv.URL
	httpClient = srv.Client()
	defer func() {
		metarBaseURL = origURL
		httpClient = origClient
	}()

	data, err := fetchMetars(map[int]string{0: "KOAK"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != body {
		t.Errorf("body mismatch:\ngot  %q\nwant %q", data, body)
	}
}

func TestFetchMetars_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	origURL := metarBaseURL
	origClient := httpClient
	metarBaseURL = srv.URL
	httpClient = srv.Client()
	defer func() {
		metarBaseURL = origURL
		httpClient = origClient
	}()

	_, err := fetchMetars(map[int]string{0: "KOAK"})
	if err == nil {
		t.Error("expected error for HTTP 500, got nil")
	}
}

func TestFetchMetars_ConnectionError(t *testing.T) {
	origURL := metarBaseURL
	origClient := httpClient
	metarBaseURL = "http://127.0.0.1:0" // nothing listening here
	httpClient = &http.Client{}
	defer func() {
		metarBaseURL = origURL
		httpClient = origClient
	}()

	_, err := fetchMetars(map[int]string{0: "KOAK"})
	if err == nil {
		t.Error("expected connection error, got nil")
	}
}
