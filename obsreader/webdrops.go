package obsreader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"path/filepath"
	"sort"
	"time"

	"github.com/meteocima/dewetra2wrf/types"
)

// This file contains a ObsReader that reads
// observations from JSON files as downloaded
// from webdrops.

type WebdropsObsReader struct{}

// ReadAll implements ObsReader
func (r WebdropsObsReader) ReadAll(dataPath string, domain types.Domain, date time.Time) ([]types.Observation, error) {

	relativeHumidity, err := readRelativeHumidity(dataPath, domain, date)
	if err != nil {
		return nil, err
	}

	temperature, err := readTemperature(dataPath, domain, date)
	if err != nil {
		return nil, err
	}

	windDirection, err := readWindDirection(dataPath, domain, date)
	if err != nil {
		return nil, err
	}

	windSpeed, err := readWindSpeed(dataPath, domain, date)
	if err != nil {
		return nil, err
	}

	precipitableWater, err := readPrecipitableWater(dataPath, domain, date)
	if err != nil {
		return nil, err
	}

	pressure, err := readPressure(dataPath, domain, date)
	if err != nil {
		return nil, err
	}

	return MergeObservations(dataPath, domain, pressure, relativeHumidity, temperature, windDirection, windSpeed, precipitableWater)
}

func readRelativeHumidity(dataPath string, domain types.Domain, date time.Time) ([]types.Result, error) {
	return readDewetraSensor(dataPath, domain, "IGROMETRO", date)
}

func readTemperature(dataPath string, domain types.Domain, date time.Time) ([]types.Result, error) {
	return readDewetraSensor(dataPath, domain, "TERMOMETRO", date)
}

func readWindDirection(dataPath string, domain types.Domain, date time.Time) ([]types.Result, error) {
	return readDewetraSensor(dataPath, domain, "DIREZIONEVENTO", date)
}

func readWindSpeed(dataPath string, domain types.Domain, date time.Time) ([]types.Result, error) {
	return readDewetraSensor(dataPath, domain, "ANEMOMETRO", date)
}

func readPrecipitableWater(dataPath string, domain types.Domain, date time.Time) ([]types.Result, error) {
	return readDewetraSensor(dataPath, domain, "PLUVIOMETRO", date)
}

func readPressure(dataPath string, domain types.Domain, date time.Time) ([]types.Result, error) {
	return readDewetraSensor(dataPath, domain, "BAROMETRO", date)
}

func readDewetraSensor(dataPath string, domain types.Domain, sensorClass string, date time.Time) ([]types.Result, error) {

	content, err := ioutil.ReadFile(filepath.Join(dataPath, sensorClass+".json"))
	if err != nil {
		return nil, err
	}
	sensorObservations := []types.Result{}
	observationLess := func(i, j int) bool {
		if sensorObservations[i].SortKey == sensorObservations[j].SortKey {
			return sensorObservations[i].At.Unix() < sensorObservations[j].At.Unix()
		}
		return sensorObservations[i].SortKey < sensorObservations[j].SortKey
	}

	data := []sensorData{}

	err = json.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}

	sensorsTable, err := openSensorsMap(dataPath, domain, sensorClass)
	if err != nil {
		return nil, err
	}

	betterTimed := map[string]types.Result{}

	for _, sens := range data {
		sensAnag, ok := sensorsTable[sens.SensorID]
		if !ok {
			continue
		}
		for idx, dateS := range sens.Timeline {
			at, err := time.Parse(time.RFC3339, dateS)
			if err != nil {
				return nil, err
			}

			sortKey := fmt.Sprintf("%s:%05f:%05f", sensAnag.Name, sensAnag.Lat, sensAnag.Lng)
			//fmt.Println(sortKey)

			betterTimedObs, ok := betterTimed[sortKey]

			dateS := date.Format("20060102 15:04")
			atS := at.Format("20060102 15:04")
			betterS := betterTimedObs.At.Format("20060102 15:04")

			_, _, _ = dateS, atS, betterS

			if !ok || math.Abs(at.Sub(date).Minutes()) < math.Abs(betterTimedObs.At.Sub(date).Minutes()) {
				sensorResult := types.Result{
					At:      at,
					Value:   sens.Values[idx],
					SortKey: sortKey,
					ID:      sens.SensorID,
				}
				betterTimed[sortKey] = sensorResult
				//sensorObservations = append(sensorObservations, sensorResult)
			}
		}
	}

	for _, sens := range betterTimed {
		sensorObservations = append(sensorObservations, sens)
	}

	sort.SliceStable(sensorObservations, observationLess)

	return sensorObservations, nil
}
