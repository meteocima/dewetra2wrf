package sensor

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// Collection is an enum that represents collections
// of meteo stations (e.g. wunderground stations)
type Collection int

// Wunderground represents wunderground stations
// DPCTrusted represents trusted italian stations
const (
	Wunderground Collection = iota
	DPCTrusted
)

// Result represent a sensor value at a point in time
type Result struct {
	SortKey string
	At      time.Time
	Value   float64
	ID      string
}

// SensorValue is
func (result Result) SensorValue() Value {
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
	TempAvg      Value
	DewptAvg     Value
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

// SortKey returns a string used to sort observations
func (obs Observation) SortKey() string {
	s := fmt.Sprintf("%s:%05f:%05f", obs.StationName, obs.Lat, obs.Lon)
	return s

}

// Constant related to Arden Buck equation
const (
	a = 6.1121 // mbar
	b = 18.678
	c = 257.14 // °C,
	d = 234.5
)

// CalculateDewpoint calcultes the dewpoint temperature using
// [Arden Buck equation[(https://en.wikipedia.org/wiki/Dew_point#Calculating_the_dew_point)
// Calculated value is stored directly in the Observation object DewptAvg
// field.
func (obs *Observation) CalculateDewpoint() {
	if obs.HumidityAvg.IsNaN() || obs.Metric.TempAvg.IsNaN() {
		obs.Metric.DewptAvg = Value(math.NaN())
		return
	}

	RH := obs.HumidityAvg.AsFloat() / 100
	T := obs.Metric.TempAvg.AsFloat()
	exp := (b - T/d) * (T / (c + T))
	gammaM := math.Log(RH * math.Pow(math.E, exp))

	obs.Metric.DewptAvg = Value((c * gammaM) / (b - gammaM))
}

// Domain is
type Domain struct {
	MinLat, MinLon, MaxLat, MaxLon float64
}
