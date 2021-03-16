package types

import (
	"strconv"
	"strings"
)

// Domain is
type Domain struct {
	MinLat, MinLon, MaxLat, MaxLon float64
}

func DomainFromS(s string) (*Domain, error) {
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
