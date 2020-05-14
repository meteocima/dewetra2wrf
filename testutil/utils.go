package testutil

import (
	"path/filepath"
	"runtime"
	"time"
)

// FixtureDir return directory of fixtures
func FixtureDir(filePath string) string {
	_, currentFilePath, _, _ := runtime.Caller(0)
	//fmt.Println("currentFilePath", currentFilePath)
	result, err := filepath.Abs(filepath.Join(currentFilePath, "../../fixtures", filePath))
	if err != nil {
		panic(err)
	}
	return result
}

func MustParse(dateS string) time.Time {
	dt, err := time.Parse("200601021504", dateS)
	if err != nil {
		panic(err)
	}
	return dt
}

func MustParseISO(dateS string) time.Time {
	dt, err := time.Parse(time.RFC3339, dateS)
	if err != nil {
		panic(err)
	}
	return dt
}
