package types

import (
	"strconv"
	"strings"
)

// Domain is a struct that represents a geographic
// area, delimited on latitude and longitued
// by max and min values.
type Domain struct {
	MinLat, MinLon, MaxLat, MaxLon float64
}

// DomainFromS returns a new Domain pointer
// accordingly to the given string, that must
// contains  MinLat,MaxLat,MinLon,MaxLon values,
// in that sequence, separated by commas and
// represented as floats.
func DomainFromS(s string) (*Domain, error) {
	if s == "" {
		return &Domain{
			MinLat: -180,
			MinLon: -90,
			MaxLat: 90,
			MaxLon: 180,
		}, nil
	}
	coords := strings.Split(s, ",")

	MinLat, err := strconv.ParseFloat(coords[0], 64)
	if err != nil {
		return nil, err
	}

	MaxLat, err := strconv.ParseFloat(coords[1], 64)
	if err != nil {
		return nil, err
	}

	MinLon, err := strconv.ParseFloat(coords[2], 64)
	if err != nil {
		return nil, err
	}

	MaxLon, err := strconv.ParseFloat(coords[3], 64)
	if err != nil {
		return nil, err
	}

	return &Domain{
		MinLat: MinLat,
		MinLon: MinLon,
		MaxLat: MaxLat,
		MaxLon: MaxLon,
	}, nil
}
