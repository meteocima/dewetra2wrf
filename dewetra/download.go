package dewetra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"path"
	"sort"
	"strconv"
	"time"

	"github.com/meteocima/wund-to-ascii/testutil"
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

// SensorResult represent a sensor value at a point in time
type SensorResult struct {
	SensorID string
	At       time.Time
	Value    float64
}

// SensorValue is
func (result SensorResult) SensorValue() SensorData {
	return SensorData(result.Value)
}

// SensorsResult represent a set of sensors value at a point in time
type SensorsResult []SensorResult

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

// SensorData is
type SensorData float64

// MarshalJSON is
func (data SensorData) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(data)) {
		return []byte("\"NaN\""), nil
	}
	return []byte(strconv.FormatFloat(float64(data), 'f', 5, 64)), nil
}

// ObservationMetric is
type ObservationMetric struct {
	TempAvg      SensorData
	DewptAvg     SensorData
	WindspeedAvg SensorData
	Pressure     SensorData
	PrecipTotal  SensorData
}

// Observation represents data for all sensor classes of
// a station at a moment in time
type Observation struct {
	StationID   string
	ObsTimeUtc  time.Time
	Lat, Lon    float64
	HumidityAvg SensorData
	WinddirAvg  SensorData
	Metric      ObservationMetric
}

func observationIsLess(this, that SensorResult) bool {
	if this.SensorID == that.SensorID {
		return this.At.Unix() < that.At.Unix()
	}
	return this.SensorID < that.SensorID
}

func minObservation(results ...SensorResult) SensorResult {
	min := SensorResult{SensorID: "ZZZZZZZZZZZZZZZZZZZZZZZZZ"}
	for _, result := range results {
		if observationIsLess(result, min) {
			min = result
		}
	}
	return min
}

// DownloadTrusted is
func DownloadTrusted(ids []string, dateFrom, dateTo time.Time) ([]Observation, error) {
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

	return matchDownloadedData(pressure, relativeHumidity, temperature, windDirection, windSpeed, precipitableWater)
}

func matchDownloadedData(pressure, relativeHumidity, temperature, windDirection, windSpeed, precipitableWater SensorsResult) ([]Observation, error) {
	pressureIdx := 0
	relativeHumidityIdx := 0
	temperatureIdx := 0
	windDirectionIdx := 0
	windSpeedIdx := 0
	precipitableWaterIdx := 0

	results := []Observation{}

	sensorsTable, err := openCompleteSensorsMap()
	if err != nil {
		return nil, err
	}

	for {

		var pressureItem SensorResult
		if len(pressure) > pressureIdx {
			pressureItem = pressure[pressureIdx]
		} else {
			pressureItem.SensorID = "ZZZZZZZZZZ"
		}

		var relativeHumidityItem SensorResult
		if len(relativeHumidity) > relativeHumidityIdx {
			relativeHumidityItem = relativeHumidity[relativeHumidityIdx]
		} else {
			relativeHumidityItem.SensorID = "ZZZZZZZZZZ"
		}

		var temperatureItem SensorResult
		if len(temperature) > temperatureIdx {
			temperatureItem = temperature[temperatureIdx]
		} else {
			temperatureItem.SensorID = "ZZZZZZZZZZ"
		}

		var windDirectionItem SensorResult
		if len(windDirection) > windDirectionIdx {
			windDirectionItem = windDirection[windDirectionIdx]
		} else {
			windDirectionItem.SensorID = "ZZZZZZZZZZ"
		}

		var windSpeedItem SensorResult
		if len(windSpeed) > windSpeedIdx {
			windSpeedItem = windSpeed[windSpeedIdx]
		} else {
			windSpeedItem.SensorID = "ZZZZZZZZZZ"
		}

		var precipitableWaterItem SensorResult
		if len(precipitableWater) > precipitableWaterIdx {
			precipitableWaterItem = precipitableWater[precipitableWaterIdx]
		} else {
			precipitableWaterItem.SensorID = "ZZZZZZZZZZ"
		}

		if relativeHumidityItem.SensorID == "ZZZZZZZZZZ" &&
			temperatureItem.SensorID == "ZZZZZZZZZZ" &&
			windDirectionItem.SensorID == "ZZZZZZZZZZ" &&
			windSpeedItem.SensorID == "ZZZZZZZZZZ" &&
			precipitableWaterItem.SensorID == "ZZZZZZZZZZ" &&
			pressureItem.SensorID == "ZZZZZZZZZZ" {
			break
		}

		minItem := minObservation(pressureItem, relativeHumidityItem, temperatureItem, windDirectionItem, windSpeedItem, precipitableWaterItem)
		station := sensorsTable[minItem.SensorID]

		currentObs := Observation{
			ObsTimeUtc:  minItem.At,
			StationID:   minItem.SensorID,
			Lat:         station.Lat,
			Lon:         station.Lon,
			HumidityAvg: SensorData(math.NaN()),
			WinddirAvg:  SensorData(math.NaN()),
			Metric: ObservationMetric{
				DewptAvg:     SensorData(math.NaN()),
				PrecipTotal:  SensorData(math.NaN()),
				Pressure:     SensorData(math.NaN()),
				TempAvg:      SensorData(math.NaN()),
				WindspeedAvg: SensorData(math.NaN()),
			},
		}

		if relativeHumidityItem.SensorID == currentObs.StationID && currentObs.ObsTimeUtc.Equal(relativeHumidityItem.At) {
			currentObs.HumidityAvg = relativeHumidityItem.SensorValue()
			relativeHumidityIdx++
		}

		if temperatureItem.SensorID == currentObs.StationID && currentObs.ObsTimeUtc.Equal(temperatureItem.At) {
			currentObs.Metric.TempAvg = temperatureItem.SensorValue()
			temperatureIdx++
		}

		if windDirectionItem.SensorID == currentObs.StationID && currentObs.ObsTimeUtc.Equal(windDirectionItem.At) {
			currentObs.WinddirAvg = windDirectionItem.SensorValue()
			windDirectionIdx++
		}

		if windSpeedItem.SensorID == currentObs.StationID && currentObs.ObsTimeUtc.Equal(windSpeedItem.At) {
			currentObs.Metric.WindspeedAvg = windSpeedItem.SensorValue()
			windSpeedIdx++
		}

		if precipitableWaterItem.SensorID == currentObs.StationID && currentObs.ObsTimeUtc.Equal(precipitableWaterItem.At) {
			currentObs.Metric.PrecipTotal = precipitableWaterItem.SensorValue()
			precipitableWaterIdx++
		}

		if pressureItem.SensorID == currentObs.StationID && currentObs.ObsTimeUtc.Equal(pressureItem.At) {
			currentObs.Metric.Pressure = precipitableWaterItem.SensorValue()
			pressureIdx++
		}

		results = append(results, currentObs)

	}

	return results, nil
}

// downloadRelativeHumidity is
func downloadRelativeHumidity(ids []string, dateFrom, dateTo time.Time) (SensorsResult, error) {
	return downloadDewetraSensor("IGROMETRO", ids, dateFrom, dateTo)
}

// downloadTemperature is
func downloadTemperature(ids []string, dateFrom, dateTo time.Time) (SensorsResult, error) {
	return downloadDewetraSensor("TERMOMETRO", ids, dateFrom, dateTo)
}

// downloadWindDirection is
func downloadWindDirection(ids []string, dateFrom, dateTo time.Time) (SensorsResult, error) {
	return downloadDewetraSensor("DIREZIONEVENTO", ids, dateFrom, dateTo)
}

// downloadWindSpeed is
func downloadWindSpeed(ids []string, dateFrom, dateTo time.Time) (SensorsResult, error) {
	return downloadDewetraSensor("ANEMOMETRO", ids, dateFrom, dateTo)
}

// downloadPrecipitableWater is
func downloadPrecipitableWater(ids []string, dateFrom, dateTo time.Time) (SensorsResult, error) {
	return downloadDewetraSensor("PLUVIOMETRO", ids, dateFrom, dateTo)
}

// downloadPrecipitableWater is
func downloadPressure(ids []string, dateFrom, dateTo time.Time) (SensorsResult, error) {
	return downloadDewetraSensor("BAROMETRO", ids, dateFrom, dateTo)
}

func downloadDewetraSensor(sensorClass string, ids []string, dateFrom, dateTo time.Time) (SensorsResult, error) {
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

	sensorObservations := SensorsResult{}
	observationLess := func(i, j int) bool {
		if sensorObservations[i].SensorID == sensorObservations[j].SensorID {
			return sensorObservations[i].At.Unix() < sensorObservations[j].At.Unix()
		}
		return sensorObservations[i].SensorID < sensorObservations[j].SensorID
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

	for _, sensor := range data {
		for idx, dateS := range sensor.Timeline {
			at, err := time.Parse("200601021504", dateS)
			if err != nil {
				return nil, err
			}

			sensAnag := sensorsTable[sensor.SensorID]
			sensorObservations = append(sensorObservations, SensorResult{
				At:       at,
				Value:    sensor.Values[idx],
				SensorID: fmt.Sprintf("%s:%05f:%05f", sensAnag.StationName, sensAnag.Lat, sensAnag.Lon),
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

	sensorsAnagContent, err := ioutil.ReadFile(path.Join(testutil.FixtureDir(".."), "data", sensorClass+".json"))
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
