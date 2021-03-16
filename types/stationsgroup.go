package types

// StationsGroup is an enum that represents
// category of meteo stations (e.g. wunderground
// stations or DPC network stations)
type StationsGroup int

const (
	// Wunderground represents wunderground stations
	Wunderground StationsGroup = iota
	// DPCTrusted represents trusted italian stations
	DPCTrusted
)
