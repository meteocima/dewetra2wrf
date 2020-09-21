package wunderground

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/meteocima/dewetra2wrf/sensor"
)

type wundObsMetric struct {
	DewptAvg     float64
	PressureMin  float64
	PressureMax  float64
	TempAvg      float64
	WindspeedAvg float64
	PrecipTotal  float64
}

type wundObs struct {
	HumidityAvg float64
	Lat         float64
	Lon         float64
	WinddirAvg  float64
	ObsTimeUtc  string
	StationID   string
	Metric      wundObsMetric
}

// Read is
func Read(dataPath string, domain sensor.Domain, date time.Time) ([]sensor.Observation, error) {
	dateDir := filepath.Join(dataPath, date.Format("2006010215"))
	files, err := ioutil.ReadDir(dateDir)
	if err != nil {
		return nil, err
	}
	observations := []sensor.Observation{}
	for _, f := range files {
		obsBuf, err := ioutil.ReadFile(filepath.Join(dateDir, f.Name()))
		if err != nil {
			return nil, err
		}
		var obs wundObs
		err = json.Unmarshal(obsBuf, &obs)
		if err != nil {
			return nil, err
		}
		if obs.Lat <= domain.MaxLat && obs.Lat >= domain.MinLat &&
			obs.Lon <= domain.MaxLon && obs.Lon >= domain.MinLon {
			dt, err := time.Parse(time.RFC3339, obs.ObsTimeUtc)
			if err != nil {
				return nil, err
			}
			resObs := sensor.Observation{
				StationID:   obs.StationID,
				StationName: obs.StationID,
				HumidityAvg: sensor.Value(obs.HumidityAvg),
				Lat:         obs.Lat,
				Lon:         obs.Lon,
				ObsTimeUtc:  dt,
				WinddirAvg:  sensor.Value(obs.WinddirAvg),
				Metric: sensor.ObservationMetric{
					WindspeedAvg: sensor.Value(obs.Metric.WindspeedAvg),
					TempAvg:      sensor.Value(obs.Metric.TempAvg),
					Pressure:     sensor.Value((obs.Metric.PressureMax + obs.Metric.PressureMin) / 2),
					PrecipTotal:  sensor.Value(obs.Metric.PrecipTotal),
				},
			}
			observations = append(observations, resObs)
		}

	}
	return observations, nil
}
