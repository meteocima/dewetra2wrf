package conversion

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/meteocima/wund-to-ascii/sensor"
)

// QC is
const QC = 0

// ERROR is
const ERROR = 0.10

func skipLines(reader *bufio.Reader, n int) {
	for i := 0; i < n; i++ {
		_, err := reader.ReadString('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			panic(err)
		}
	}
}

func str(s string, len int) string {
	strFmt := fmt.Sprintf("%%-%ds", len)
	return fmt.Sprintf(strFmt, s)
}

func integer(i int, len int) string {
	intS := fmt.Sprintf("%d", i)
	strFmt := fmt.Sprintf("%%%ds", len)
	return fmt.Sprintf(strFmt, intS)
}

func num(f sensor.Value, len float64) string {
	if f.IsNaN() {
		f = -888888.0
	}

	strFmt := fmt.Sprintf("%% %sf", strconv.FormatFloat(float64(len), 'f', -1, 64))
	return fmt.Sprintf(strFmt, f)
}

func space(n int) string {
	return strings.Repeat(" ", n)
}

func date(dt time.Time) string {
	return dt.Format("2006-01-02_15:04:05")
}

func dataQCError(data string) string {
	return data +
		integer(QC, 4) +
		num(ERROR, 7.2)
}

func dataQCError3(data string) string {
	return data +
		integer(QC, 4) +
		num(ERROR, 7.3)
}

//INFO  = PLATFORM, DATE, NAME, LEVELS, LATITUDE, LONGITUDE, ELEVATION, ID.
//SRFC  = SLP, PW (DATA,QC,ERROR).
//EACH  = PRES, SPEED, DIR, HEIGHT, TEMP, DEW PT, HUMID (DATA,QC,ERROR)*LEVELS.
//INFO_FMT = (A12,1X,A19,1X,A40,1X,I6,3(F12.3,11X),6X,A40)
//SRFC_FMT = (F12.3,I4,F7.2,F12.3,I4,F7.3)
//EACH_FMT = (3(F12.3,I4,F7.2),11X,3(F12.3,I4,F7.2),11X,3(F12.3,I4,F7.2))

// ToWRFDA is
func ToWRFDA(obs sensor.Observation) string {
	firstLine :=
		str("FM-12 SYNOP", 12) +
			" " +
			date(obs.ObsTimeUtc) +
			" " +
			str("XXXXXX", 40) +
			" " +
			integer(1, 6) +
			num(sensor.Value(obs.Lat), 12.3) +
			space(11) +
			num(sensor.Value(obs.Lon), 12.3) +
			space(11) +
			num(sensor.Value(obs.Elevation), 12.3) +
			space(11) +
			space(6) +
			str("XXXXXX", 40)

	surfaceLevelPressure := sensor.Value(0.0)
	secondLine :=
		dataQCError(num(surfaceLevelPressure, 12.3)) +
			dataQCError3(num(obs.Metric.PrecipTotal, 12.3))

	thirstLine :=
		dataQCError(num(obs.Metric.Pressure, 12.3)) +
			dataQCError(num(obs.Metric.WindspeedAvg, 12.3)) +
			dataQCError(num(obs.WinddirAvg, 12.3)) +
			space(11) +
			dataQCError(num(sensor.Value(obs.Elevation), 12.3)) +
			dataQCError(num(obs.Metric.TempAvg, 12.3)) +
			dataQCError(num(obs.Metric.DewptAvg, 12.3)) +
			space(11) +
			dataQCError(num(obs.HumidityAvg, 12.3)) +
			dataQCError(num(0.0, 12.3)) +
			dataQCError(num(0.0, 12.3))

	return firstLine + "\n" + secondLine + "\n" + thirstLine
}
