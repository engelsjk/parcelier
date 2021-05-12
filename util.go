package parcelier

import (
	"fmt"
	"io/ioutil"
	"os"
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

func DirExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

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

func SaveGeoJSON(filePath string, data []byte) error {
	return ioutil.WriteFile(filePath, data, 0644)
}

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

func GetFeatureCollection(data []byte) (*geojson.FeatureCollection, error) {
	return geojson.UnmarshalFeatureCollection(data)
}

func MatchParcelsCount(tilePath, parcelPath string) (bool, error) {
	b, err := LoadFile(tilePath)
	if err != nil {
		return false, fmt.Errorf("unable to load tile file")
	}
	tile, err := GetFeature(b)
	if err != nil {
		return false, fmt.Errorf("unable to load tile feature")
	}
	b, err = LoadFile(parcelPath)
	if err != nil {
		return false, fmt.Errorf("unable to load parcels file")
	}
	parcels, err := GetFeatureCollection(b)
	if err != nil {
		return false, fmt.Errorf("unable to load parcels feature collection")
	}
	numParcels, ok := tile.Properties["num_parcels"].(float64)
	if !ok {
		return false, fmt.Errorf("tile num_parcels property is not available")
	}
	if int(numParcels) != len(parcels.Features) {
		return false, fmt.Errorf("tile/parcels count doesn't match")
	}
	return true, nil
}
