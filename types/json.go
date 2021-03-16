package types

import (
	"fmt"
	"time"
)

// Result represent a sensor value at a point in time
// as read from CIMA webdrops webapi.
type Result struct {
	SortKey string
	At      time.Time
	Value   float64
	ID      string
}

// ObservationMetric is
type ObservationMetric struct {
	TempAvg Value
	//DewptAvg     Value
	WindspeedAvg Value
	Pressure     Value
	PrecipTotal  Value
}

// Observation represents data for all sensor classes of
// a station at a moment in time
type Observation struct {
	Elevation   float64
	StationID   string
	StationName string
	ObsTimeUtc  time.Time
	Lat, Lon    float64
	HumidityAvg Value
	WinddirAvg  Value
	Metric      ObservationMetric
}

type WundObsMetric struct {
	DewptAvg     float64
	PressureMin  float64
	PressureMax  float64
	TempAvg      float64
	WindspeedAvg float64
	PrecipTotal  float64
}

type WundObs struct {
	HumidityAvg float64
	Lat         float64
	Lon         float64
	WinddirAvg  float64
	ObsTimeUtc  string
	StationID   string
	Metric      WundObsMetric
}

// SortKey returns a string used to sort observations
func (obs Observation) SortKey() string {
	s := fmt.Sprintf("%s:%05f:%05f", obs.StationName, obs.Lat, obs.Lon)
	return s

}
