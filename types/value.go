package types

import (
	"math"
	"strconv"
)

// Value is a type that represents a value
// as read from a weather station sensor.
// It has the particularity that, when the
// type is JSON unmarshaled, JSON string
// equals to "NaN" are converted to math.NaN.
type Value float64

// AsFloat returns the Value casted back to a float64
func (data Value) AsFloat() float64 {
	return float64(data)
}

// IsNaN check if the Value is math.NaN()
func (data Value) IsNaN() bool {
	return math.IsNaN(float64(data))
}

// NaN returns a new Value containing math.NaN()
func NaN() Value {
	return Value(math.NaN())
}

// MarshalJSON implements json.Marshaler
func (data Value) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(data)) {
		return []byte("\"NaN\""), nil
	}
	return []byte(strconv.FormatFloat(float64(data), 'f', 5, 64)), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (data *Value) UnmarshalJSON(buff []byte) error {
	buffS := string(buff)
	if buffS == "\"NaN\"" {
		*data = Value(math.NaN())
		return nil
	}
	if buffS == "null" {
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
