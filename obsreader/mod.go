// Package obsreader contains an `ObsReader` interface
// for types that can read list of types.Observation.
//
// It also contains three implementations of
// the interface:
//
//  * WebdropsObsReader    - reads observations from a set of JSON files that follows the dewetra observations format
//  * WundCurrentObsReader - reads observations from a set of JSON files as archived from the WSDN CIMA process.
//  * WundHistObsReader    - reads observations from a set of JSON files as returned from the Wunderground API service.
package obsreader
