package parcelier

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/engelsjk/arcgis2geojson"
	"github.com/paulmach/orb/geojson"
)

const (
	ZoomLimit int = 25
)

const (
	StatusAPIError                = "api error"
	StatusJSONConversionError     = "json conversion error"
	StatusGeoJSONError            = "geojson error"
	StatusSaveError               = "unable to save error"
	StatusSaveSucceeded           = "save succeeded"
	StatusError                   = "error"
	StatusTileFileExistsWithMatch = "tile file exists (parcel count matches)"
	StatusTileNoParcels           = "tile no parcels"
	StatusTileOK                  = "tile ok"
	StatusTileExceedsParcelLimit  = "tile exceeds parcel limit"
	StatusUnknown                 = "unknown"
)

type Options struct {
	URL              string
	Agent            string
	ParcelsPath      string
	Update           bool
	TimeWait         int
	ParcelLimit      int
	SpatialReference string
	Format           string
	ParcelID         string
	ParcelKey        string
	TilesPath        string
	Verbose          bool
	VeryVerbose      bool
}

type Parcels struct {
	FeatureCollection *geojson.FeatureCollection
	Count             int
}

func Run(tiler *Tiler, options Options) {

	api := NewAPI(options.URL, options.Agent)
	api.verbose = options.Verbose
	api.veryVerbose = options.VeryVerbose

	limiter := time.Tick(time.Millisecond * time.Duration(options.TimeWait))

	for i := 0; i < tiler.NumTiles; i++ {
		tile := tiler.GetTileAtIndex(i)
		<-limiter
		GetParcels(api, tile, tiler.Zoom, options)
	}
}

// GetParcels ...
func GetParcels(api *API, tile Tile, zoom int, opts Options) {

	parcels, status, err := GetParcelsFromTile(api, tile, opts)

	baseLog := fmt.Sprintf("status: %s | tile: %s | parcels: %d", status, tile.GetTileString(), parcels.Count)

	switch status {
	case StatusError:
		if opts.Verbose {
			fmt.Printf("%s | %s\n", baseLog, err)
			break
		}
		fmt.Println(baseLog)
	case StatusTileFileExistsWithMatch:
		fmt.Printf("%s | %s\n", baseLog, "skipping")
	case StatusTileNoParcels:
		fmt.Println(baseLog)
	case StatusTileOK:
		saveLog, _ := SaveParcels(tile, parcels, opts)
		fmt.Printf("%s | saving...%s\n", baseLog, saveLog)
		if opts.TilesPath != "" {
			saveLog, _ = SaveTile(tile, parcels.Count, opts)
			fmt.Printf("(option) saving tile...%s\n", saveLog)
		}
	case StatusTileExceedsParcelLimit:
		fmt.Printf("%s | %s\n", baseLog, "iterating")

		newZoom := zoom + 1
		if newZoom >= ZoomLimit {
			fmt.Printf("%s | zoom limit %d exceeded!\n", baseLog, ZoomLimit)
			break
		}

		fmt.Printf("rerunning tile %s at zoom %d...\n", tile.GetTileString(), newZoom)
		newTiler := NewTiler(tile.Bound().ToPolygon(), newZoom)
		Run(newTiler, opts)
	case StatusUnknown:
		fmt.Println(baseLog)
	default:
		fmt.Printf("%s | %s\n", baseLog, err)
	}
}

// GetParcelsFromTile queries the API for parcels using the given tile.
func GetParcelsFromTile(api *API, tile Tile, opts Options) (*Parcels, string, error) {

	// todo: check if json response w/ error message
	// {"error":{"code":400,"message":"Failed to execute query.","details":[]}}

	tilePath := tile.GetTilePath(opts.TilesPath)

	if FileExists(tilePath) && !opts.Update {
		parcelPath := tile.GetParcelPath(opts.ParcelsPath)
		ok, numParcels, _ := ParcelsCountMatch(tilePath, parcelPath)
		if ok {
			return &Parcels{Count: numParcels}, StatusTileFileExistsWithMatch, nil
		}
	}

	extent := tile.GetExtentString()

	var format string
	switch strings.ToUpper(opts.Format) {
	case "GEOJSON":
		format = "geoJSON"
	case "JSON":
		format = "JSON"
	}

	queryParams := map[string]string{
		"orderByFields":  opts.ParcelID,
		"geometry":       extent,
		"geometryType":   "esriGeometryEnvelope",
		"returnGeometry": "true",
		"where":          "",
		"f":              format,
		"outfields":      "*",
		"spatialRel":     "esriSpatialRelEnvelopeIntersects",
		"inSR":           opts.SpatialReference,
		"outSR":          opts.SpatialReference,
	}

	b, err := api.Get(queryParams)
	if err != nil {
		return nil, StatusAPIError, err
	}

	if format == "JSON" {
		b, err = arcgis2geojson.Convert(b, opts.ParcelID)
		if err != nil {
			return nil, StatusJSONConversionError, err
		}
	}

	featureCollection, err := GetFeatureCollection(b)
	if err != nil {
		return nil, StatusGeoJSONError, err
	}

	parcels := &Parcels{
		FeatureCollection: featureCollection,
		Count:             len(featureCollection.Features),
	}

	if parcels.Count == 0 {
		return parcels, StatusTileNoParcels, nil
	}
	if parcels.Count < opts.ParcelLimit {
		return parcels, StatusTileOK, nil
	}
	if parcels.Count >= opts.ParcelLimit {
		return parcels, StatusTileExceedsParcelLimit, nil
	}

	return parcels, StatusUnknown, nil
}

// ParcelsCountMatch compares a tile and parcel file for a matching parcel count.
func ParcelsCountMatch(tilePath, parcelPath string) (bool, int, error) {
	b, err := LoadFile(tilePath)
	if err != nil {
		return false, 0, fmt.Errorf("unable to load tile file")
	}
	tile, err := GetFeature(b)
	if err != nil {
		return false, 0, fmt.Errorf("unable to load tile feature")
	}
	b, err = LoadFile(parcelPath)
	if err != nil {
		return false, 0, fmt.Errorf("unable to load parcels file")
	}
	parcels, err := GetFeatureCollection(b)
	if err != nil {
		return false, 0, fmt.Errorf("unable to load parcels feature collection")
	}
	var numParcels int
	np, ok := tile.Properties["num_parcels"].(float64)
	if !ok {
		return false, int(np), fmt.Errorf("tile num_parcels property is not available")
	}
	numParcels = int(np)
	if numParcels != len(parcels.Features) {
		return false, numParcels, fmt.Errorf("tile/parcels count doesn't match")
	}
	return true, numParcels, nil
}

// SaveParcels saves a FeatureCollection of parcels for the given tile.
func SaveParcels(tile Tile, parcels *Parcels, opts Options) (string, error) {
	b, err := json.MarshalIndent(parcels.FeatureCollection, "", " ")
	if err != nil {
		return StatusGeoJSONError, err
	}
	filePath := tile.GetParcelPath(opts.ParcelsPath)
	err = SaveGeoJSON(filePath, b)
	if err != nil {
		return StatusSaveError, err
	}
	return StatusSaveSucceeded, nil
}

// SaveTile saves a Feature for the given tile.
func SaveTile(tile Tile, numParcels int, opts Options) (string, error) {
	f := geojson.NewFeature(tile.Bound().ToPolygon())
	f.Properties["extent"] = tile.GetTileString()
	f.Properties["num_parcels"] = numParcels
	b, err := f.MarshalJSON()
	if err != nil {
		return StatusGeoJSONError, err
	}
	filePath := tile.GetTilePath(opts.TilesPath)
	err = SaveGeoJSON(filePath, b)
	if err != nil {
		return StatusSaveError, err
	}
	return StatusSaveSucceeded, nil
}

// PrintInfo prints the boundary file and count of covering tiles.
func PrintInfo(boundaryFile string, tiler *Tiler) {
	fmt.Printf("running boundary %s\n", path.Base(boundaryFile))
	fmt.Printf("covering %d tiles at zoom %d\n", tiler.NumTiles, tiler.Zoom)
}

// PrintDone prints the number of tiles processed.
func PrintDone(tiler *Tiler) {
	fmt.Printf("done! %d tiles processed\n", len(tiler.Set))
}
