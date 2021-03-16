// package elevations contains a single function
// that returns elevation at specified latitude:longitude
// according to an orografy dataset contained at
// ~/.dewetra2wrf/orog.nc
package elevations

import (
	"os"
	"path"

	"github.com/meteocima/dewetra2wrf/elevations/internal/ncdf"
)

type elevationsFile struct {
	xs, ys []float64
	zs     []int32
}

var elev *elevationsFile = openElevationsFile()

func openElevationsFile() *elevationsFile {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	orog := path.Join(home, ".dewetra2wrf", "orog.nc")
	f := ncdf.File{}
	f.Open(orog)
	defer f.Close()
	x := f.Var("x")
	y := f.Var("y")
	z := f.Var("z")
	e := &elevationsFile{
		xs: x.ValuesFloat64(),
		ys: y.ValuesFloat64(),
		zs: z.ValuesInt32(),
	}
	if f.Error() != nil {
		panic(err)
	}

	return e

}

// GetFromCoord returns elevation at specified lat:lon
func GetFromCoord(lat, lon float64) float64 {
	ypos := int((0.5 + lat/180) * float64(len(elev.ys)))
	xpos := int((0.5 + lon/360) * float64(len(elev.xs)))

	val := float64(elev.zs[xpos+ypos*len(elev.xs)])
	return val
}
