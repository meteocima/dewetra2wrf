package testutil

import (
	"path/filepath"
	"runtime"
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
