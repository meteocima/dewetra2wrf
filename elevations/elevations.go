// package elevations contains a single function
// that returns elevation at specified latitude:longitude
// according to an orografy dataset contained at
// ~/.dewetra2wrf/orog.nc
package elevations

import (
	"math"
	"os"
	"path"

	//"github.com/RobinRCM/sklearn/interpolate"
	"github.com/meteocima/dewetra2wrf/elevations/internal/ncdf"
)

//var interp func(x float64, y float64) float64

type elevationsFile struct {
	xs, ys []float64
	zs     []float64
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
		zs: z.ValuesFloat64(),
	}

	//interp = interpolate.Interp2d(e.xs, e.ys, e.zs)

	if f.Error() != nil {
		panic(err)
	}

	return e

}

// GetFromCoord returns elevation at specified lat:lon
func GetFromCoord(lat, lon float64) float64 {
	// coordinates for borders of our DEM file
	minlon := elev.xs[0]
	maxlon := elev.xs[len(elev.xs)-1]
	maxlat := elev.ys[0]
	minlat := elev.ys[len(elev.ys)-1]

	// translate all lat:lon coordinates of -minlon, -minlat
	// in order to obtain a coordinate system with origin on leftmost
	// and topmost pixel of our DEM file.
	maxlonTr := maxlon - minlon // this is also the width in degrees of the DEM
	maxlatTr := maxlat - minlat // this is also the height in degrees of the DEM
	latTr := lat - minlat
	lonTr := lon - minlon

	// resolution in pixel of our DEM
	yres := float64(len(elev.ys))
	xres := float64(len(elev.xs))

	// having our DEM uniform resolution, we have
	// latTr / maxlatTr == ypos / yres
	// lonTr / maxlonTr == xpos / xres
	// from this equation, we can calculate xpos and ypos
	// since y axis is inverted, we should
	// invert ypos too
	yposF := float64(len(elev.ys)) - latTr/maxlatTr*yres
	//fmt.Println("yposF", yposF)

	xposF := lonTr / maxlonTr * xres
	//fmt.Println("xposF", xposF)

	xpos := int(math.Round(xposF))
	ypos := int(math.Round(yposF))

	//fmt.Println("xpos", xpos, "ypos", ypos)
	val := elev.zs[xpos+ypos*len(elev.xs)]

	// Missing values means lat:lon fall on sea
	// so return 0 as altitude
	if math.IsNaN(val) || val == -9999 {
		return 0
	}

	return val
}
