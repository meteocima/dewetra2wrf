package obsreader

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/meteocima/dewetra2wrf/elevations"
	"github.com/meteocima/dewetra2wrf/types"
)

// This file contains a ObsReader that reads
// observations from JSON files as returned
// from wunderground 'current' web API.

type WundCurrentObsReader struct{}

// ReadAll implements ObsReader
func (r WundCurrentObsReader) ReadAll(dataPath string, domain types.Domain, date time.Time) ([]types.Observation, error) {
	dateDir := filepath.Join(dataPath, date.Format("2006010215"))
	files, err := ioutil.ReadDir(dateDir)
	if err != nil {
		return nil, err
	}
	observations := []types.Observation{}

	for _, f := range files {
		obsBuf, err := ioutil.ReadFile(filepath.Join(dateDir, f.Name()))
		if err != nil {
			return nil, err
		}
		var obs types.WundObs
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
			resObs := types.Observation{
				Elevation:   elevations.GetFromCoord(obs.Lat, obs.Lon),
				StationID:   obs.StationID,
				StationName: obs.StationID,
				HumidityAvg: types.Value(obs.HumidityAvg),
				Lat:         obs.Lat,
				Lon:         obs.Lon,
				ObsTimeUtc:  dt,
				WinddirAvg:  types.Value(obs.WinddirAvg),
				Metric: types.ObservationMetric{
					WindspeedAvg: types.Value(obs.Metric.WindspeedAvg),
					TempAvg:      types.Value(obs.Metric.TempAvg),
					Pressure:     types.Value((obs.Metric.PressureMax + obs.Metric.PressureMin) / 2),
					PrecipTotal:  types.Value(obs.Metric.PrecipTotal),
				},
			}
			// convert temperatures from °celsius to °kelvin
			resObs.Metric.TempAvg += 273.15
			// convert wind speed from km/h into m/s
			resObs.Metric.WindspeedAvg *= 0.277778
			// convert pressure from mbar into Pa
			resObs.Metric.Pressure *= 100

			observations = append(observations, resObs)
		}

	}
	return observations, nil
}
