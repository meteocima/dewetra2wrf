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

// SensorValue returns the value
// contained in this Result instance, or NaN()
// if not value is present.
func (result Result) SensorValue() Value {
	if result.Value == -9998 {
		return NaN()
	}
	return Value(result.Value)
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

// ObservationMetric contains a subset of values
// contained in an Observation
type ObservationMetric struct {
	TempAvg      Value
	DewptAvg     Value
	WindspeedAvg Value
	Pressure     Value
	PrecipTotal  Value
	PressureMin  Value
	PressureMax  Value
}

// SortKey returns a string used to sort observations
func (obs Observation) SortKey() string {
	s := fmt.Sprintf("%s:%05f:%05f", obs.StationName, obs.Lat, obs.Lon)
	return s

}
