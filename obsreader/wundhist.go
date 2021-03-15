package obsreader

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"path/filepath"
	"time"

	"github.com/meteocima/dewetra2wrf/elevations"
	"github.com/meteocima/dewetra2wrf/sensor"
)

// This file contains a ObsReader that reads
// observations from JSON files as returned
// from wunderground 'historical' web API.

type WundHistObsReader struct{}

// ReadAll implements ObsReader
func (r WundHistObsReader) ReadAll(dataPath string, domain sensor.Domain, date time.Time) ([]sensor.Observation, error) {
	dateDir := filepath.Join(dataPath, date.Format("2006010215"))
	files, err := ioutil.ReadDir(dateDir)
	if err != nil {
		return nil, err
	}
	observations := []sensor.Observation{}
	elevations, err := elevations.OpenElevationsFile(dataPath)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		obsBuf, err := ioutil.ReadFile(filepath.Join(dateDir, f.Name()))
		if err != nil {
			return nil, err
		}
		var obsList struct {
			Observations []sensor.WundObs
		}

		err = json.Unmarshal(obsBuf, &obsList)
		if err != nil {
			return nil, err
		}

		if len(obsList.Observations) == 0 {
			continue
		}
		var obs sensor.WundObs = obsList.Observations[0]

		if obs.Lat <= domain.MaxLat && obs.Lat >= domain.MinLat &&
			obs.Lon <= domain.MaxLon && obs.Lon >= domain.MinLon {

			/*



			 */
			minDeltaMin := 30.0
			for _, o := range obsList.Observations {
				dtObs, err := time.Parse(time.RFC3339, o.ObsTimeUtc)
				if err != nil {
					return nil, err
				}

				delta := math.Abs(date.Sub(dtObs).Minutes())
				if delta < minDeltaMin {
					minDeltaMin = delta
					obs = o
				}
				if delta == 0.0 {
					break
				}
			}

			/*




			 */
			dt, err := time.Parse(time.RFC3339, obs.ObsTimeUtc)
			if err != nil {
				return nil, err
			}
			resObs := sensor.Observation{
				Elevation:   elevations.GetElevation(obs.Lat, obs.Lon),
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
