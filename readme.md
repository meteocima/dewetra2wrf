# Dewetra to WRF

[![Coverage](https://coveralls.io/repos/github/meteocima/dewetra2wrf/badge.svg?branch=master)](https://coveralls.io/github/meteocima/dewetra2wrf?branch=master) [![CI](https://github.com/meteocima/dewetra2wrf/actions/workflows/go.yml/badge.svg)](https://github.com/meteocima/dewetra2wrf/actions/workflows/go.yml) [![Docs](https://pkg.go.dev/badge/github.com/meteocima/dewetra2wrf.svg)](https://pkg.go.dev/github.com/meteocima/dewetra2wrf) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


This module is the forefront of a framework
of modules that can be used to convert weather
stations observations in various format into ascii
WRF format.


## Installation

The module use go-netcdf in order to read a netcdf
containing world orography data.

In order to use it, you need the developer version of the
library provided by your distribution installed.

On ubuntu you can install it with:

```bash
sudo apt install libnetcdf-dev
```

You can download the orography file from 
https://zenodo.org/record/4607436/files/orog.nc


## Command line usage

This module implements a console command
that can be used to convert observation
as returned from webdrops API to ascii
WRF format.

Usage of `d2w`:

```
d2w [options]
Options:
  -date string
        date and hour of the data to download [YYYYMMDDHH]
  -domain string
        domain to filter stations to download [MinLat,MaxLat,MinLon,MaxLon]
  -format string
        format of input files (DEWETRA or WUNDERGROUND) (default ".")
  -input string
        where to read input files (default ".")
  -outfile string
        where to save converted file (default "./out")
```