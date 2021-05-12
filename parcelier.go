package parcelier

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/engelsjk/arcgis2geojson"
	"github.com/paulmach/orb/geojson"
)

const (
	ZoomLimit int = 25
)

const (
	StatusAPIError               = "status: api error"
	StatusJSONConversionError    = "status: json conversion error"
	StatusGeoJSONError           = "status: geojson error"
	StatusSaveError              = "status: unable to save error"
	StatusSaveSucceeded          = "status: save succeeded"
	StatusError                  = "status: error"
	StatusTileFileExists         = "status: tile file exists"
	StatusTileNoParcels          = "status: tile no parcels"
	StatusTileOK                 = "status: tile ok"
	StatusTileExceedsParcelLimit = "status: tile exceeds parcel limit"
	StatusUnknown                = "status: unknown"
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
	NumFeatures       int
}

func Run(tiler *Tiler, options Options) {

	api := NewAPI(options.URL, options.Agent)
	api.verbose = options.Verbose
	api.veryVerbose = options.VeryVerbose

	limiter := time.Tick(time.Millisecond * time.Duration(options.TimeWait))

	for i := 0; i < tiler.NumTiles; i++ {

		tile, err := tiler.GetTileAtIndex(i)
		if err != nil {
			fmt.Println(err)
			continue
		}

		tilePath := tile.GetTilePath(options.TilesPath)

		if FileExists(tilePath) && !options.Update {
			parcelPath := tile.GetParcelPath(options.ParcelsPath)
			if ok, err := MatchParcelsCount(tilePath, parcelPath); ok {
				fmt.Printf("%s | tile: %s | %s...skipping\n", StatusTileFileExists, tile.GetTileString(), "parcels count match")
			} else {
				fmt.Printf("%s | tile: %s | %s...skipping\n", StatusTileFileExists, tile.GetTileString(), err)
			}
			continue
		}

		<-limiter
		GetParcels(api, tile, tiler.Zoom, options)
	}
}

func GetParcels(api *API, tile Tile, zoom int, opts Options) {
	var numParcels int

	parcels, status, err := GetParcelsFromTile(api, tile, opts)

	if parcels == nil {
		numParcels = -1
	} else {
		numParcels = parcels.NumFeatures
	}

	baseLog := fmt.Sprintf("%s | tile: %s | parcels: %d", status, tile.GetTileString(), numParcels)

	switch status {
	case StatusError:
		fmt.Println(baseLog)
		return
	case StatusTileFileExists:
		fmt.Printf("%s | %s\n", baseLog, filepath.Base(tile.GetTilePath(opts.TilesPath)))
		return
	case StatusTileNoParcels:
		fmt.Println(baseLog)
		return
	case StatusTileOK:
		logSave, _ := SaveParcels(tile, parcels, opts)
		fmt.Printf("%s | saving...%s\n", baseLog, logSave)
		if opts.TilesPath != "" {
			filePath := tile.GetTilePath(opts.TilesPath)
			f := geojson.NewFeature(tile.Bound().ToPolygon())
			f.Properties["extent"] = tile.GetTileString()
			f.Properties["num_parcels"] = parcels.NumFeatures
			b, _ := f.MarshalJSON()
			SaveGeoJSON(filePath, b)
		}
		return
	case StatusTileExceedsParcelLimit:
		fmt.Printf("%s | %s\n", baseLog, "")
		newZoom := zoom + 1
		if newZoom < ZoomLimit {
			fmt.Printf("rerunning tile %s at zoom %d...\n", tile.GetTileString(), newZoom)
			newTiler := NewTiler(tile.Bound().ToPolygon(), newZoom)
			Run(newTiler, opts)
		} else {
			fmt.Printf("%s | zoom limit %d exceeded!\n", baseLog, ZoomLimit)
			return
		}
	case StatusUnknown:
		fmt.Println(baseLog)
		return
	default:
		// todo: handle all Status*Error returns from GetParcelsFromTile
		fmt.Printf("%s | %s\n", status, err)
		return
	}
}

func GetParcelsFromTile(api *API, tile Tile, options Options) (*Parcels, string, error) {

	extent := fmt.Sprintf("%s", tile.GetExtentString())

	if FileExists(tile.GetTilePath(options.ParcelsPath)) && !options.Update {
		return nil, StatusTileFileExists, nil
	}

	var f string
	switch strings.ToUpper(options.Format) {
	case "GEOJSON":
		f = "geoJSON"
	case "JSON":
		f = "JSON"
	}

	queryParams := map[string]string{
		"orderByFields":  options.ParcelID,
		"geometry":       extent,
		"geometryType":   "esriGeometryEnvelope",
		"returnGeometry": "true",
		"where":          "",
		"f":              f,
		"outfields":      "*",
		"spatialRel":     "esriSpatialRelEnvelopeIntersects",
		"inSR":           options.SpatialReference,
		"outSR":          options.SpatialReference,
	}

	b, err := api.Get(queryParams)
	if err != nil {
		return nil, StatusAPIError, err
	}

	// todo: check if json response w/ error message
	// {"error":{"code":400,"message":"Failed to execute query.","details":[]}}

	if f == "JSON" {
		b, err = arcgis2geojson.Convert(b, options.ParcelID)
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
		NumFeatures:       len(featureCollection.Features),
	}

	if parcels.NumFeatures == 0 {
		return parcels, StatusTileNoParcels, nil
	}
	if parcels.NumFeatures < options.ParcelLimit {
		return parcels, StatusTileOK, nil
	}
	if parcels.NumFeatures >= options.ParcelLimit {
		return parcels, StatusTileExceedsParcelLimit, nil
	}
	return parcels, StatusUnknown, nil
}

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

func PrintInfo(boundaryFile string, tiler *Tiler) {
	fmt.Printf("running boundary %s\n", path.Base(boundaryFile))
	fmt.Printf("getting %d tiles at zoom %d\n", tiler.NumTiles, tiler.Zoom)
}

func PrintDone(tiler *Tiler) {
	fmt.Printf("done! %d tiles processed\n", len(tiler.Set))
}
