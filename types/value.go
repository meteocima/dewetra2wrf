package types

import (
	"math"
	"strconv"
)

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
