package sensor

import (
	"fmt"
	"math"
	"strconv"
	"time"
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
