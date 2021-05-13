package parcelier

import (
	"io/ioutil"
	"os"
	"path/filepath"

	geojson "github.com/paulmach/orb/geojson"
)

// FileExists ...
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirExists ...
func DirExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// LoadFile ...
func LoadFile(filename string) ([]byte, error) {

	var err error
	if !filepath.IsAbs(filename) {
		filename, err = filepath.Abs(filename)
		if err != nil {
			return nil, err
		}
	}

	file, err := os.OpenFile(filename, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	file.Close()
	return b, nil
}

// SaveGeoJSON ...
func SaveGeoJSON(filePath string, data []byte) error {
	return ioutil.WriteFile(filePath, data, 0644)
}

// GetFeature ...
func GetFeature(b []byte) (*geojson.Feature, error) {
	fc, err := GetFeatureCollection(b)
	if err == nil {
		return fc.Features[0], nil
	}
	f, err := geojson.UnmarshalFeature(b)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// GetFeatureCollection ...
func GetFeatureCollection(data []byte) (*geojson.FeatureCollection, error) {
	return geojson.UnmarshalFeatureCollection(data)
}
