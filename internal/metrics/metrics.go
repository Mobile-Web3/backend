package metrics

import (
	"time"

	"github.com/paulbellamy/ratecounter"
)

var (
	RpsCounter    = ratecounter.NewRateCounter(time.Minute * 60)
	ErrorsCounter = ratecounter.NewRateCounter(time.Minute * 60)
	PanicsCounter = ratecounter.NewRateCounter(time.Minute * 60)
)
