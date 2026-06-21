package main

import (
	"bufio"
	"flag"
	"fmt"
	"image/color"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/finack/twinkle/internal/config"
	"github.com/finack/twinkle/internal/display"
	"github.com/finack/twinkle/internal/metardata"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var catNames = []string{"VFR", "MVFR", "IFR", "LIFR"}

const (
	maxKt      = 44.0
	sweepSteps = 100
	sweepDelay = 30 * time.Millisecond
	holdTime   = 8 * time.Second
)

// vfrWindSteps shows all map LEDs as VFR at each key wind speed, stepping on Enter.
func vfrWindSteps(leds *display.Leds, c config.Config) {
	steps := []float64{0, 10, 15, 20, 25, 32, 40}
	labels := []string{
		"0 kt — calm (pure VFR green)",
		"10 kt — at low threshold (no change yet)",
		"15 kt — 33% shift toward yellow-green",
		"20 kt — 67% shift toward yellow-green",
		"25 kt — at high threshold (full yellow-green)",
		"32 kt — yellow-green + 19% white",
		"40 kt — yellow-green + 40% white (max)",
	}

	scanner := bufio.NewScanner(os.Stdin)
	for i, kt := range steps {
		col := metardata.FlightColor("VFR", kt, c.WindLowKt, c.WindHighKt)
		for ledNum := range c.Leds {
			leds.Display(ledNum, col)
		}
		leds.Ws.Render()
		fmt.Printf("\n[%d/%d] %s\n       R=%d G=%d B=%d\nPress Enter for next...",
			i+1, len(steps), labels[i], col.R, col.G, col.B)
		scanner.Scan()
	}
}

// showGradient fills the strip with a wind-speed gradient: each category gets an
// equal slice of LEDs, ranging from calm (left) to stormy (right).
func showGradient(leds *display.Leds, c config.Config) {
	ledsPerCat := c.LedCount / len(catNames)
	for catIdx, cat := range catNames {
		for pos := 0; pos < ledsPerCat; pos++ {
			kt := float64(pos) / float64(ledsPerCat-1) * maxKt
			leds.Display(catIdx*ledsPerCat+pos, metardata.FlightColor(cat, kt, c.WindLowKt, c.WindHighKt))
		}
	}
	for i := len(catNames) * ledsPerCat; i < c.LedCount; i++ {
		leds.Display(i, color.RGBA{})
	}
	leds.Ws.Render()
}

// sweepCategory ramps all LEDs through 0→maxKt→0 for a single flight category.
func sweepCategory(leds *display.Leds, c config.Config, cat string) {
	ramp := func(start, end int) {
		for step := start; step != end; step += sign(end - start) {
			kt := float64(step) / sweepSteps * maxKt
			col := metardata.FlightColor(cat, kt, c.WindLowKt, c.WindHighKt)
			for i := 0; i < c.LedCount; i++ {
				leds.Display(i, col)
			}
			leds.Ws.Render()
			time.Sleep(sweepDelay)
		}
	}
	ramp(0, sweepSteps)
	time.Sleep(500 * time.Millisecond)
	ramp(sweepSteps, 0)
	time.Sleep(500 * time.Millisecond)
}

func sign(n int) int {
	if n < 0 {
		return -1
	}
	return 1
}

func main() {
	configFile := flag.String("config", "config.yaml", "Path to configuration file")
	steps := flag.Bool("steps", false, "Step through VFR wind states on the real map layout")
	flag.Parse()

	c := config.GetConfig(configFile)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	leds, err := display.New(c.Brightness, c.LedCount)
	if err != nil {
		log.Fatal().Err(err).Caller().Msg("Could not setup LEDs")
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		leds.Clear()
		os.Exit(0)
	}()

	if *steps {
		log.Info().
			Float64("lowKt", c.WindLowKt).
			Float64("highKt", c.WindHighKt).
			Msg("VFR wind steps — real map positions")
		vfrWindSteps(leds, c)
		leds.Clear()
		return
	}

	ledsPerCat := c.LedCount / len(catNames)
	log.Info().
		Int("ledsPerCategory", ledsPerCat).
		Float64("maxKt", maxKt).
		Float64("lowKt", c.WindLowKt).
		Float64("highKt", c.WindHighKt).
		Msg("Wind color demo")
	log.Info().Msgf("Gradient layout: VFR[0-%d] MVFR[%d-%d] IFR[%d-%d] LIFR[%d-%d] — left=calm right=stormy",
		ledsPerCat-1, ledsPerCat, 2*ledsPerCat-1, 2*ledsPerCat, 3*ledsPerCat-1, 3*ledsPerCat, 4*ledsPerCat-1)

	for {
		log.Info().Msg("Phase 1: gradient")
		showGradient(leds, c)
		time.Sleep(holdTime)

		log.Info().Msg("Phase 2: sweep")
		for _, cat := range catNames {
			log.Info().Str("category", cat).Msg("Sweep")
			sweepCategory(leds, c, cat)
		}
	}
}
