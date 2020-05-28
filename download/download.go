package download

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
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
	ID                  string
	StationName         string
	Lon, Lat, Elevation float64
}

func observationIsLess(this, that sensor.Result) bool {
	if this.SortKey == that.SortKey {
		return this.At.Unix() < that.At.Unix()
	}
	return this.SortKey < that.SortKey
}

func minObservation(results ...sensor.Result) sensor.Result {
	min := sensor.Result{SortKey: "zzzzzzzzzzzzzzzzzzzzzzzzz"}
	for _, result := range results {
		if observationIsLess(result, min) {
			min = result
		}
	}
	return min
}

type byName struct {
	ids   []string
	table map[string]sensorAnag
}

// Len is the number of elements in the collection.
func (bn byName) Len() int {
	return len(bn.ids)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (bn byName) Less(i, j int) bool {
	idI := bn.ids[i]
	idJ := bn.ids[j]
	return bn.table[idI].StationName < bn.table[idJ].StationName
}

// Swap swaps the elements with indexes i and j.
func (bn byName) Swap(i, j int) {
	save := bn.ids[i]
	bn.ids[i] = bn.ids[j]
	bn.ids[j] = save
}

func getSensorsIds(dataPath string, domain sensor.Domain, sensorClass string) ([]string, error) {
	sensorsTable := map[string]sensorAnag{}
	err := fillSensorsMap(dataPath, sensorClass, sensorsTable)
	if err != nil {
		return nil, err
	}

	ids := []string{}
	for _, sens := range sensorsTable {
		if sens.Lat >= domain.MinLat && sens.Lat <= domain.MaxLat &&
			sens.Lon >= domain.MinLon && sens.Lon <= domain.MaxLon {
			ids = append(ids, sens.ID)
		}
	}

	sort.Sort(byName{ids, sensorsTable})

	return ids, nil
}

// AllSensors is
func AllSensors(dataPath string, domain sensor.Domain, date time.Time) ([]sensor.Observation, error) {
	ids, err := getSensorsIds(dataPath, domain, "IGROMETRO")
	if err != nil {
		return nil, err
	}
	relativeHumidity, err := downloadRelativeHumidity(dataPath, ids, date)
	if err != nil {
		return nil, err
	}

	ids, err = getSensorsIds(dataPath, domain, "TERMOMETRO")
	if err != nil {
		return nil, err
	}
	temperature, err := downloadTemperature(dataPath, ids, date)
	if err != nil {
		return nil, err
	}

	ids, err = getSensorsIds(dataPath, domain, "DIREZIONEVENTO")
	if err != nil {
		return nil, err
	}
	windDirection, err := downloadWindDirection(dataPath, ids, date)
	if err != nil {
		return nil, err
	}

	ids, err = getSensorsIds(dataPath, domain, "ANEMOMETRO")
	if err != nil {
		return nil, err
	}
	windSpeed, err := downloadWindSpeed(dataPath, ids, date)
	if err != nil {
		return nil, err
	}

	ids, err = getSensorsIds(dataPath, domain, "PLUVIOMETRO")
	if err != nil {
		return nil, err
	}
	precipitableWater, err := downloadPrecipitableWater(dataPath, ids, date)
	if err != nil {
		return nil, err
	}

	ids, err = getSensorsIds(dataPath, domain, "BAROMETRO")
	if err != nil {
		return nil, err
	}
	pressure, err := downloadPressure(dataPath, ids, date)
	if err != nil {
		return nil, err
	}

	return MatchDownloadedData(dataPath, pressure, relativeHumidity, temperature, windDirection, windSpeed, precipitableWater)
}

// MatchDownloadedData is
func MatchDownloadedData(dataPath string, pressure, relativeHumidity, temperature, windDirection, windSpeed, precipitableWater []sensor.Result) ([]sensor.Observation, error) {
	pressureIdx := 0
	relativeHumidityIdx := 0
	temperatureIdx := 0
	windDirectionIdx := 0
	windSpeedIdx := 0
	precipitableWaterIdx := 0

	results := []sensor.Observation{}

	sensorsTable, err := openCompleteSensorsMap(dataPath)
	if err != nil {
		return nil, err
	}

	for {

		var pressureItem sensor.Result
		if len(pressure) > pressureIdx {
			pressureItem = pressure[pressureIdx]
		} else {
			pressureItem.SortKey = "zzzzzzzzzz"
		}

		var relativeHumidityItem sensor.Result
		if len(relativeHumidity) > relativeHumidityIdx {
			relativeHumidityItem = relativeHumidity[relativeHumidityIdx]
		} else {
			relativeHumidityItem.SortKey = "zzzzzzzzzz"
		}

		var temperatureItem sensor.Result
		if len(temperature) > temperatureIdx {
			temperatureItem = temperature[temperatureIdx]
		} else {
			temperatureItem.SortKey = "zzzzzzzzzz"
		}

		var windDirectionItem sensor.Result
		if len(windDirection) > windDirectionIdx {
			windDirectionItem = windDirection[windDirectionIdx]
		} else {
			windDirectionItem.SortKey = "zzzzzzzzzz"
		}

		var windSpeedItem sensor.Result
		if len(windSpeed) > windSpeedIdx {
			windSpeedItem = windSpeed[windSpeedIdx]
		} else {
			windSpeedItem.SortKey = "zzzzzzzzzz"
		}

		var precipitableWaterItem sensor.Result
		if len(precipitableWater) > precipitableWaterIdx {
			precipitableWaterItem = precipitableWater[precipitableWaterIdx]
		} else {
			precipitableWaterItem.SortKey = "zzzzzzzzzz"
		}

		if relativeHumidityItem.SortKey == "zzzzzzzzzz" &&
			temperatureItem.SortKey == "zzzzzzzzzz" &&
			windDirectionItem.SortKey == "zzzzzzzzzz" &&
			windSpeedItem.SortKey == "zzzzzzzzzz" &&
			precipitableWaterItem.SortKey == "zzzzzzzzzz" &&
			pressureItem.SortKey == "zzzzzzzzzz" {
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
			Elevation:   station.Elevation,
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
			currentObs.Metric.Pressure = pressureItem.SensorValue()
			pressureIdx++
		}

		// formula for dewpoint calculation must be applied with
		// temperature stil in celsius
		currentObs.CalculateDewpoint()

		// convert temperatures from °celsius to °kelvin
		currentObs.Metric.DewptAvg += 273.15
		currentObs.Metric.TempAvg += 273.15

		results = append(results, currentObs)

	}

	return results, nil
}

func downloadRelativeHumidity(dataPath string, ids []string, date time.Time) ([]sensor.Result, error) {
	return downloadDewetraSensor(dataPath, "IGROMETRO", ids, date)
}

func downloadTemperature(dataPath string, ids []string, date time.Time) ([]sensor.Result, error) {
	return downloadDewetraSensor(dataPath, "TERMOMETRO", ids, date)
}

func downloadWindDirection(dataPath string, ids []string, date time.Time) ([]sensor.Result, error) {
	return downloadDewetraSensor(dataPath, "DIREZIONEVENTO", ids, date)
}

func downloadWindSpeed(dataPath string, ids []string, date time.Time) ([]sensor.Result, error) {
	return downloadDewetraSensor(dataPath, "ANEMOMETRO", ids, date)
}

func downloadPrecipitableWater(dataPath string, ids []string, date time.Time) ([]sensor.Result, error) {
	return downloadDewetraSensor(dataPath, "PLUVIOMETRO", ids, date)
}

func downloadPressure(dataPath string, ids []string, date time.Time) ([]sensor.Result, error) {
	return downloadDewetraSensor(dataPath, "BAROMETRO", ids, date)
}

func downloadDewetraSensor(dataPath string, sensorClass string, ids []string, date time.Time) ([]sensor.Result, error) {
	url := fmt.Sprintf("%s/drops_sensors/serie", baseURL)

	dateFrom := date.Add(-time.Minute * 30)
	dateTo := date.Add(time.Minute * 30)

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

	sensorsTable, err := openSensorsMap(dataPath, sensorClass)
	if err != nil {
		return nil, err
	}

	betterTimed := map[string]sensor.Result{}

	for _, sens := range data {
		for idx, dateS := range sens.Timeline {
			at, err := time.Parse("200601021504", dateS)
			if err != nil {
				return nil, err
			}

			at = at.Add(time.Hour * 2)

			sensAnag := sensorsTable[sens.SensorID]
			sortKey := fmt.Sprintf("%s:%05f:%05f", sensAnag.StationName, sensAnag.Lat, sensAnag.Lon)

			betterTimedObs, ok := betterTimed[sortKey]

			dateS := date.Format("20060102 15:04")
			atS := at.Format("20060102 15:04")
			betterS := betterTimedObs.At.Format("20060102 15:04")

			_, _, _ = dateS, atS, betterS

			if !ok || math.Abs(at.Sub(date).Minutes()) < math.Abs(betterTimedObs.At.Sub(date).Minutes()) {
				sensorResult := sensor.Result{
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

func openSensorsMap(dataPath string, sensorClass string) (map[string]sensorAnag, error) {
	sensorsTable := map[string]sensorAnag{}

	err := fillSensorsMap(dataPath, sensorClass, sensorsTable)
	if err != nil {
		return nil, err
	}

	return sensorsTable, nil
}

func getElevation(sensor sensorAnag) (float64, error) {
	url := fmt.Sprintf("https://api.airmap.com/elevation/v1/ele/?points=%6f,%6f", sensor.Lat, sensor.Lon)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-API-Key", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjcmVkZW50aWFsX2lkIjoiY3JlZGVudGlhbHx2YkJZOEs2dTAzRFdYZ1NaZDRvbkVodjJhMG9SIiwiYXBwbGljYXRpb25faWQiOiJhcHBsaWNhdGlvbnxEbzhFV1hKQzBOOEU4elVwcW45dkxUbFBYUE1SIiwib3JnYW5pemF0aW9uX2lkIjoiZGV2ZWxvcGVyfGE2Tk9LNUJ0S3ZBeHZkY2V3SzJrdmZ3ZUo5cUUiLCJpYXQiOjE1ODk3OTA2MTd9.dzqy2VbQtmHrf4sKwlb3S0PdLiqGS4ms4LFvKeWmMkY")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return 0, fmt.Errorf("HTTP response %d", res.StatusCode)
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	var data struct {
		Data []float64
	}

	err = json.Unmarshal(content, &data)
	if err != nil {
		return 0, err
	}

	return data.Data[0], nil
}

func readElevationsFromFile(dataPath string) (map[string]float64, error) {
	elevFile := path.Join(dataPath, "elevations.json")
	elevations := map[string]float64{}

	elevationsContent, err := ioutil.ReadFile(elevFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err == nil {
		err = json.Unmarshal(elevationsContent, &elevations)
		if err != nil {
			return nil, err
		}
	}

	return elevations, nil
}

func saveElevationsToFile(dataPath string, elevations map[string]float64) error {
	buff, err := json.MarshalIndent(elevations, " ", " ")
	if err != nil {
		return err
	}
	elevFile := path.Join(dataPath, "elevations.json")

	return ioutil.WriteFile(elevFile, buff, os.FileMode(0644))

}

func fillSensorsMap(dataPath string /*, domain sensor.Domain*/, sensorClass string, sensorsTable map[string]sensorAnag) error {
	sensorsAnag := []sensorAnag{}
	///*testutil.FixtureDir(".."),*/ "../data"
	sensorsAnagContent, err := ioutil.ReadFile(path.Join(dataPath, sensorClass+".json"))
	if err != nil {
		return err
	}

	err = json.Unmarshal(sensorsAnagContent, &sensorsAnag)
	if err != nil {
		return err
	}

	elevations, err := readElevationsFromFile(dataPath)
	if err != nil {
		return err
	}

	for n, sensor := range sensorsAnag {
		stKey := fmt.Sprintf("%f:%f", sensor.Lat, sensor.Lon)
		elevation, ok := elevations[stKey]
		if !ok {
			fmt.Println(n, stKey)
			elevation, err := getElevation(sensor)
			if err != nil {
				fmt.Println(err)
				continue
			}
			elevations[stKey] = elevation
			err = saveElevationsToFile(dataPath, elevations)
			if err != nil {
				return err
			}
		}

		sensor.Elevation = elevation
		sensorsTable[sensor.ID] = sensor
	}

	return nil
}

func DownloadAllSensorsTables(dataPath string, collection sensor.Collection) error {
	err := downloadSensorsTable(dataPath, collection, "IGROMETRO")
	if err != nil {
		return err
	}

	err = downloadSensorsTable(dataPath, collection, "TERMOMETRO")
	if err != nil {
		return err
	}

	err = downloadSensorsTable(dataPath, collection, "DIREZIONEVENTO")
	if err != nil {
		return err
	}

	err = downloadSensorsTable(dataPath, collection, "ANEMOMETRO")
	if err != nil {
		return err
	}

	err = downloadSensorsTable(dataPath, collection, "PLUVIOMETRO")
	if err != nil {
		return err
	}

	err = downloadSensorsTable(dataPath, collection, "BAROMETRO")
	if err != nil {
		return err
	}

	return nil
}

func downloadSensorsTable(dataPath string, collection sensor.Collection, sensorClass string) error {
	url := fmt.Sprintf("%s/drops_sensors/anag/%s/%s", baseURL, sensorClass, collection.Key())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("HTTP response %d", res.StatusCode)
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/%s.json", dataPath, sensorClass)
	return ioutil.WriteFile(filename, content, os.FileMode(0644))

}

func openCompleteSensorsMap(dataPath string) (map[string]sensorAnag, error) {
	sensorsTable := map[string]sensorAnag{}

	err := fillSensorsMap(dataPath, "IGROMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}

	err = fillSensorsMap(dataPath, "TERMOMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}

	err = fillSensorsMap(dataPath, "DIREZIONEVENTO", sensorsTable)
	if err != nil {
		return nil, err
	}

	err = fillSensorsMap(dataPath, "ANEMOMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}

	err = fillSensorsMap(dataPath, "PLUVIOMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}

	err = fillSensorsMap(dataPath, "BAROMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}

	return sensorsTable, nil
}
