package download

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"path"
	"sort"
	"time"

	"github.com/meteocima/wund-to-ascii/sensor"
)

const baseURL = "http://dds.cimafoundation.org/dds/rest"
const username = "admin"
const password = "geoDDS2013"

/*

> * wind speed
ANEMOMETRO
> * wind direction
DIREZIONEVENTO
> * dewpoint temperature
Non esiste, puoi calcolarla
> * temperature
TERMOMETRO
> * relative humidity
IGROMETRO
> * precipitable water
PLUVIOMETRO

*/

type sensorReqBody struct {
	SensorClass string   `json:"sensorClass"`
	From        string   `json:"from"`
	To          string   `json:"to"`
	Ids         []string `json:"ids"`
}

type sensorData struct {
	SensorID string
	Timeline []string
	Values   []float64
}

type sensorAnag struct {
	ID          string
	StationName string
	Lon, Lat    float64
}

func observationIsLess(this, that sensor.Result) bool {
	if this.SortKey == that.SortKey {
		return this.At.Unix() < that.At.Unix()
	}
	return this.SortKey < that.SortKey
}

func minObservation(results ...sensor.Result) sensor.Result {
	min := sensor.Result{SortKey: "ZZZZZZZZZZZZZZZZZZZZZZZZZ"}
	for _, result := range results {
		if observationIsLess(result, min) {
			min = result
		}
	}
	return min
}

// AllSensors is
func AllSensors(ids []string, dateFrom, dateTo time.Time) ([]sensor.Observation, error) {
	relativeHumidity, err := downloadRelativeHumidity(ids, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	temperature, err := downloadTemperature(ids, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	windDirection, err := downloadWindDirection(ids, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	windSpeed, err := downloadWindSpeed(ids, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	precipitableWater, err := downloadPrecipitableWater(ids, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	pressure, err := downloadPressure(ids, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	return MatchDownloadedData(pressure, relativeHumidity, temperature, windDirection, windSpeed, precipitableWater)
}

// MatchDownloadedData is
func MatchDownloadedData(pressure, relativeHumidity, temperature, windDirection, windSpeed, precipitableWater []sensor.Result) ([]sensor.Observation, error) {
	pressureIdx := 0
	relativeHumidityIdx := 0
	temperatureIdx := 0
	windDirectionIdx := 0
	windSpeedIdx := 0
	precipitableWaterIdx := 0

	results := []sensor.Observation{}

	sensorsTable, err := openCompleteSensorsMap()
	if err != nil {
		return nil, err
	}

	for {

		var pressureItem sensor.Result
		if len(pressure) > pressureIdx {
			pressureItem = pressure[pressureIdx]
		} else {
			pressureItem.SortKey = "ZZZZZZZZZZ"
		}

		var relativeHumidityItem sensor.Result
		if len(relativeHumidity) > relativeHumidityIdx {
			relativeHumidityItem = relativeHumidity[relativeHumidityIdx]
		} else {
			relativeHumidityItem.SortKey = "ZZZZZZZZZZ"
		}

		var temperatureItem sensor.Result
		if len(temperature) > temperatureIdx {
			temperatureItem = temperature[temperatureIdx]
		} else {
			temperatureItem.SortKey = "ZZZZZZZZZZ"
		}

		var windDirectionItem sensor.Result
		if len(windDirection) > windDirectionIdx {
			windDirectionItem = windDirection[windDirectionIdx]
		} else {
			windDirectionItem.SortKey = "ZZZZZZZZZZ"
		}

		var windSpeedItem sensor.Result
		if len(windSpeed) > windSpeedIdx {
			windSpeedItem = windSpeed[windSpeedIdx]
		} else {
			windSpeedItem.SortKey = "ZZZZZZZZZZ"
		}

		var precipitableWaterItem sensor.Result
		if len(precipitableWater) > precipitableWaterIdx {
			precipitableWaterItem = precipitableWater[precipitableWaterIdx]
		} else {
			precipitableWaterItem.SortKey = "ZZZZZZZZZZ"
		}

		if relativeHumidityItem.SortKey == "ZZZZZZZZZZ" &&
			temperatureItem.SortKey == "ZZZZZZZZZZ" &&
			windDirectionItem.SortKey == "ZZZZZZZZZZ" &&
			windSpeedItem.SortKey == "ZZZZZZZZZZ" &&
			precipitableWaterItem.SortKey == "ZZZZZZZZZZ" &&
			pressureItem.SortKey == "ZZZZZZZZZZ" {
			break
		}

		minItem := minObservation(pressureItem, relativeHumidityItem, temperatureItem, windDirectionItem, windSpeedItem, precipitableWaterItem)
		station := sensorsTable[minItem.ID]

		currentObs := sensor.Observation{
			ObsTimeUtc:  minItem.At,
			StationID:   station.ID,
			StationName: station.StationName,
			Lat:         station.Lat,
			Lon:         station.Lon,
			HumidityAvg: sensor.Value(math.NaN()),
			WinddirAvg:  sensor.Value(math.NaN()),
			Metric: sensor.ObservationMetric{
				DewptAvg:     sensor.Value(math.NaN()),
				PrecipTotal:  sensor.Value(math.NaN()),
				Pressure:     sensor.Value(math.NaN()),
				TempAvg:      sensor.Value(math.NaN()),
				WindspeedAvg: sensor.Value(math.NaN()),
			},
		}

		if relativeHumidityItem.SortKey == currentObs.SortKey() && currentObs.ObsTimeUtc.Equal(relativeHumidityItem.At) {
			currentObs.HumidityAvg = relativeHumidityItem.SensorValue()
			relativeHumidityIdx++
		}

		if temperatureItem.SortKey == currentObs.SortKey() && currentObs.ObsTimeUtc.Equal(temperatureItem.At) {
			currentObs.Metric.TempAvg = temperatureItem.SensorValue()
			temperatureIdx++
		}

		if windDirectionItem.SortKey == currentObs.SortKey() && currentObs.ObsTimeUtc.Equal(windDirectionItem.At) {
			currentObs.WinddirAvg = windDirectionItem.SensorValue()
			windDirectionIdx++
		}

		if windSpeedItem.SortKey == currentObs.SortKey() && currentObs.ObsTimeUtc.Equal(windSpeedItem.At) {
			currentObs.Metric.WindspeedAvg = windSpeedItem.SensorValue()
			windSpeedIdx++
		}

		if precipitableWaterItem.SortKey == currentObs.SortKey() && currentObs.ObsTimeUtc.Equal(precipitableWaterItem.At) {
			currentObs.Metric.PrecipTotal = precipitableWaterItem.SensorValue()
			precipitableWaterIdx++
		}

		if pressureItem.SortKey == currentObs.SortKey() && currentObs.ObsTimeUtc.Equal(pressureItem.At) {
			currentObs.Metric.Pressure = precipitableWaterItem.SensorValue()
			pressureIdx++
		}

		results = append(results, currentObs)

	}

	return results, nil
}

func downloadRelativeHumidity(ids []string, dateFrom, dateTo time.Time) ([]sensor.Result, error) {
	return downloadDewetraSensor("IGROMETRO", ids, dateFrom, dateTo)
}

func downloadTemperature(ids []string, dateFrom, dateTo time.Time) ([]sensor.Result, error) {
	return downloadDewetraSensor("TERMOMETRO", ids, dateFrom, dateTo)
}

func downloadWindDirection(ids []string, dateFrom, dateTo time.Time) ([]sensor.Result, error) {
	return downloadDewetraSensor("DIREZIONEVENTO", ids, dateFrom, dateTo)
}

func downloadWindSpeed(ids []string, dateFrom, dateTo time.Time) ([]sensor.Result, error) {
	return downloadDewetraSensor("ANEMOMETRO", ids, dateFrom, dateTo)
}

func downloadPrecipitableWater(ids []string, dateFrom, dateTo time.Time) ([]sensor.Result, error) {
	return downloadDewetraSensor("PLUVIOMETRO", ids, dateFrom, dateTo)
}

func downloadPressure(ids []string, dateFrom, dateTo time.Time) ([]sensor.Result, error) {
	return downloadDewetraSensor("BAROMETRO", ids, dateFrom, dateTo)
}

func downloadDewetraSensor(sensorClass string, ids []string, dateFrom, dateTo time.Time) ([]sensor.Result, error) {
	url := fmt.Sprintf("%s/drops_sensors/serie", baseURL)

	sensorReq := sensorReqBody{
		SensorClass: sensorClass,
		From:        dateFrom.Format("200601021504"),
		To:          dateTo.Format("200601021504"),
		Ids:         ids,
	}

	reqBody, err := json.Marshal(sensorReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP response %d", res.StatusCode)
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	sensorObservations := []sensor.Result{}
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

	sensorsTable, err := openSensorsMap(sensorClass)
	if err != nil {
		return nil, err
	}

	for _, sens := range data {
		for idx, dateS := range sens.Timeline {
			at, err := time.Parse("200601021504", dateS)
			if err != nil {
				return nil, err
			}

			sensAnag := sensorsTable[sens.SensorID]
			sensorObservations = append(sensorObservations, sensor.Result{
				At:      at,
				Value:   sens.Values[idx],
				SortKey: fmt.Sprintf("%s:%05f:%05f", sensAnag.StationName, sensAnag.Lat, sensAnag.Lon),
				ID:      sens.SensorID,
			})
		}
	}

	sort.SliceStable(sensorObservations, observationLess)

	return sensorObservations, nil
}

func openSensorsMap(sensorClass string) (map[string]sensorAnag, error) {
	sensorsTable := map[string]sensorAnag{}

	err := fillSensorsMap(sensorClass, sensorsTable)
	if err != nil {
		return nil, err
	}

	return sensorsTable, nil
}

func fillSensorsMap(sensorClass string, sensorsTable map[string]sensorAnag) error {
	sensorsAnag := []sensorAnag{}

	sensorsAnagContent, err := ioutil.ReadFile(path.Join( /*testutil.FixtureDir(".."),*/ "../data", sensorClass+".json"))
	if err != nil {
		return err
	}

	err = json.Unmarshal(sensorsAnagContent, &sensorsAnag)
	if err != nil {
		return err
	}

	for _, sensor := range sensorsAnag {
		sensorsTable[sensor.ID] = sensor
	}

	return nil
}

func openCompleteSensorsMap() (map[string]sensorAnag, error) {
	sensorsTable := map[string]sensorAnag{}

	err := fillSensorsMap("IGROMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}

	err = fillSensorsMap("TERMOMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}

	err = fillSensorsMap("DIREZIONEVENTO", sensorsTable)
	if err != nil {
		return nil, err
	}

	err = fillSensorsMap("ANEMOMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}

	err = fillSensorsMap("PLUVIOMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}

	err = fillSensorsMap("BAROMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}

	return sensorsTable, nil
}
