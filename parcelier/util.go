package parcelier

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	geojson "github.com/paulmach/orb/geojson"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func LoadFile(filename string) ([]byte, error) {

	var err error
	if !filepath.IsAbs(filename) {
		filename, err = filepath.Abs(filename)
		if err != nil {
			return nil, err
		}
	}

	fmt.Printf("opening %s\n", path.Base(filename))

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

func SaveGeoJSON(filePath string, data []byte) error {
	return ioutil.WriteFile(filePath, data, 0644)
}

func GetFeature(data []byte) *geojson.Feature {
	// Need to better handle FC vs F!
	fc, err := geojson.UnmarshalFeatureCollection(data)
	if err == nil {
		return fc.Features[0]
	}
	f, err := geojson.UnmarshalFeature(data)
	if err != nil {
		return nil
	}
	return f
}

func GetFeatureCollection(data []byte) (*geojson.FeatureCollection, error) {
	return geojson.UnmarshalFeatureCollection(data)
}
