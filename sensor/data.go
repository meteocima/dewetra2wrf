package sensor

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// Collection is an enum that represents
// category of meteo stations (e.g. wunderground
// stations or DPC network stations)
type Collection int

const (
	// Wunderground represents wunderground stations
	Wunderground Collection = iota
	// DPCTrusted represents trusted italian stations
	DPCTrusted
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

// Value is
type Value float64

// AsFloat is
func (data Value) AsFloat() float64 {
	return float64(data)
}

// IsNaN is
func (data Value) IsNaN() bool {
	return math.IsNaN(float64(data))
}

func NaN() Value {
	return Value(math.NaN())
}

// MarshalJSON is
func (data Value) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(data)) {
		return []byte("\"NaN\""), nil
	}
	return []byte(strconv.FormatFloat(float64(data), 'f', 5, 64)), nil
}

// UnmarshalJSON is
func (data *Value) UnmarshalJSON(buff []byte) error {
	buffS := string(buff)
	if buffS == "\"NaN\"" {
		*data = Value(math.NaN())
		return nil
	}
	val, err := strconv.ParseFloat(buffS, 64)
	if err != nil {
		return err
	}
	*data = Value(val)
	return nil
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
