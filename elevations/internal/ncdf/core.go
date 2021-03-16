// package ncdf implements a netcdf file readers
// that can cache variables.
package ncdf

import (
	"fmt"
	"math"
	"time"

	"github.com/fhs/go-netcdf/netcdf"
)

// Attrs ...
type Attrs map[string]string

// Vars ...
type Vars map[string]*Variable

// File ...
type File struct {
	Attrs Attrs
	ds    *netcdf.Dataset
	err   error
	vars  Vars
}

// Variable ...
type Variable struct {
	Name     string
	Len      uint64
	Min, Max float64
	Attrs    Attrs
	Type     string
	variable netcdf.Var
	file     *File
	values   interface{}
}

// OpenFile ...
func OpenFile(filename string) *File {
	f := &File{}
	f.Open(filename)
	return f
}

// Error ...
func (data *File) Error() error {
	return data.err
}

// IsOpen ...
func (data *File) IsOpen() bool {
	return data.ds != nil
}

// Close ...
func (data *File) Close() {
	if data.err != nil {
		return
	}
	if data.ds == nil {
		panic("File closed")
	}

	data.err = data.ds.Close()
	data.ds = nil
}

// AllVars ...
func (data *File) AllVars() Vars {

	if data.err != nil {
		return nil
	}

	res := Vars{}

	nvars := 0
	nvars, data.err = data.ds.NVars()
	if data.err != nil {
		return nil
	}

	for i := 0; i < nvars; i++ {
		v := data.ds.VarN(i)
		name := ""
		name, data.err = v.Name()
		res[name] = data.Var(name)

		if data.err != nil {
			return nil
		}
	}

	return res
}

type withAttr interface {
	AttrN(n int) (a netcdf.Attr, err error)
	NAttrs() (n int, err error)
}

func (data *File) readAttributes(attrs withAttr) Attrs {

	if data.err != nil {
		return nil
	}

	nFileAttrs, err := attrs.NAttrs()
	if err != nil {
		data.err = err
		return nil
	}

	attributes := make(Attrs, nFileAttrs)

	for i := 0; i < nFileAttrs; i++ {

		attr, err := attrs.AttrN(i)
		if err != nil {
			data.err = err
			return nil
		}

		t, err := attr.Type()
		if err != nil {
			data.err = err
			return nil
		}

		if t != netcdf.CHAR {
			continue
		}
		l, err := attr.Len()
		if err != nil {
			data.err = err
			return nil
		}

		content := make([]byte, l)
		err = attr.ReadBytes(content)
		if err != nil {
			data.err = err
			return nil
		}
		attributes[attr.Name()] = string(content)
	}

	return attributes
}

// Attrib ...
func (data *File) Attrib(name string) string {
	if data.err != nil {
		return ""
	}
	if data.ds == nil {
		data.err = fmt.Errorf("no file opened")
		return ""
	}
	if data.Attrs == nil {
		data.Attrs = data.readAttributes(data.ds)
	}

	return data.Attrs[name]
}

// AllAttribs ...
func (data *File) AllAttribs() map[string]string {
	if data.err != nil {
		return map[string]string{}
	}
	if data.ds == nil {
		data.err = fmt.Errorf("no file opened")
		return map[string]string{}
	}
	if data.Attrs == nil {
		data.Attrs = data.readAttributes(data.ds)
	}

	return data.Attrs
}

// Var ...
func (data *File) Var(name string) *Variable {
	if data.err != nil {
		return &Variable{file: data}
	}
	if data.vars == nil {
		data.vars = Vars{}
	}
	if cached, ok := data.vars[name]; ok {
		return cached
	}

	if data.ds == nil {
		data.err = fmt.Errorf("no file opened")
		return &Variable{file: data}
	}

	res := Variable{
		file: data,
		Name: name,
	}

	res.variable, data.err = data.ds.Var(name)
	if data.err != nil {
		return &Variable{file: data}
	}

	var t netcdf.Type
	t, data.err = res.variable.Type()
	if data.err != nil {
		return &Variable{file: data}
	}

	res.Type = t.String()

	res.Len, data.err = res.variable.Len()
	if data.err != nil {
		return &Variable{file: data}
	}
	data.vars[res.Name] = &res
	return &res
}

// AllAttribs ...
func (v *Variable) AllAttribs() map[string]string {
	if v.file.err != nil {
		return map[string]string{}
	}
	if v.file.ds == nil {
		v.file.err = fmt.Errorf("no file opened")
		return map[string]string{}
	}
	if v.Attrs == nil {
		v.Attrs = v.file.readAttributes(v.variable)
	}

	return v.file.Attrs
}

// Attrib ...
func (v *Variable) Attrib(name string) string {
	if v.file.err != nil {
		return ""
	}
	if v.file.ds == nil {
		v.file.err = fmt.Errorf("no file opened")
		return ""
	}
	if v.Attrs == nil {
		v.Attrs = v.file.readAttributes(v.variable)
	}

	return v.Attrs[name]
}

func (v *Variable) String() string {
	return v.Name
}

// ValuesFloat64 ...
func (v *Variable) ValuesFloat64() []float64 {
	if v.file.err != nil {
		return []float64{}
	}
	if v.values != nil {
		return v.values.([]float64)
	}

	varval := make([]float64, v.Len)
	v.values = varval

	err := v.variable.ReadFloat64s(varval)
	if err != nil {
		v.file.err = err
		return nil
	}

	return varval
}

// ValuesInt32 ...
func (v *Variable) ValuesInt32() []int32 {
	if v.file.err != nil {
		return []int32{}
	}
	if v.values != nil {
		return v.values.([]int32)
	}

	varval := make([]int32, v.Len)
	v.values = varval

	err := v.variable.ReadInt32s(varval)
	if err != nil {
		v.file.err = err
		return nil
	}

	return varval
}

// TimeRepresentation ...
type TimeRepresentation int

// TimeRepresentation values
const (
	HoursFromY2K TimeRepresentation = iota
	SecondsFrom1970
)

var y2K = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

// ValuesTime ...
func (v *Variable) ValuesTime() []time.Time {
	if v.file.err != nil {
		return []time.Time{}
	}
	if v.values != nil {
		return v.values.([]time.Time)
	}

	rawVal := v.ValuesFloat64()
	if v.file.err != nil {
		return []time.Time{}
	}
	timeInstants := make([]time.Time, len(rawVal))

	for idx, hoursFromY2K := range rawVal {
		timeInstants[idx] = y2K.Add(time.Duration(hoursFromY2K) * time.Hour)
	}
	v.values = timeInstants
	return timeInstants
}

// ValuesFloat32 ...
func (v *Variable) ValuesFloat32() []float32 {
	if v.file.err != nil {
		return []float32{}
	}
	if v.values != nil {
		return v.values.([]float32)
	}
	varval := make([]float32, v.Len)
	v.values = varval

	err := v.variable.ReadFloat32s(varval)
	if err != nil {
		v.file.err = err
		return nil
	}

	return varval
}

// Dims ...
func (v *Variable) Dims() (dims []uint64) {
	dims = []uint64{}
	if v.file.err != nil {
		return
	}

	dims, v.file.err = v.variable.LenDims()
	if v.file.err != nil {
		return
	}

	return
}

// Max32 ...
func (v *Variable) Max32() float32 {
	return float32(v.Max)
}

// Min32 ...
func (v *Variable) Min32() float32 {
	return float32(v.Min)
}

// ComputeBoundaries32 ...
func (v *Variable) ComputeBoundaries32() {
	if v.file.err != nil {
		return
	}
	values := v.ValuesFloat32()

	max := float32(-math.MaxFloat32)
	min := float32(math.MaxFloat32)
	for _, val := range values {
		if val > max {
			max = val
		}
		if val < min {
			min = val
		}
	}
	v.Max = float64(max)
	v.Min = float64(min)
}

// Open ...
func (data *File) Open(filename string) {
	if data.err != nil {
		return
	}
	if data.ds != nil {
		data.err = fmt.Errorf("file already open")
		return
	}

	ds, err := netcdf.OpenFile(filename, netcdf.NOWRITE)
	data.ds, data.err = &ds, err
	if err != nil {
		data.ds = nil
	}
}
