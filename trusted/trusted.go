package trusted

import (
	"time"

	"github.com/meteocima/wund-to-ascii/conversion"
	"github.com/meteocima/wund-to-ascii/sensor"

	"github.com/meteocima/wund-to-ascii/download"
)

func DownloadAndConvert(dataPath string, domain sensor.Domain, from, to time.Time) ([]string, error) {
	download.DownloadAllSensorsTables(dataPath, sensor.DPCTrusted)

	sensorsObservations, err := download.AllSensors(dataPath, domain, from, to)
	if err != nil {
		return nil, err
	}

	results := make([]string, len(sensorsObservations))
	for i, result := range sensorsObservations {
		results[i] = conversion.ToWRFDA(result)
	}

	return results, nil
}
