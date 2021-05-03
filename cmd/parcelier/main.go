package main

import (
	"log"

	"github.com/engelsjk/parcelier"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	boundaryFilepath = kingpin.Flag("boundary", "boundary filepath").Default("").Short('b').String()
	extent           = kingpin.Flag("extent", "extent tile (z/x/y)").Default("").String()
	url              = kingpin.Flag("url", "esri url").Default("").String()
	zoom             = kingpin.Flag("zoom", "initial zoom").Default("13").Short('z').Int()
	outputFilepath   = kingpin.Flag("output", "parcel output dir").Default(".").Short('o').String()

	saveTilePath = kingpin.Flag("tiles", "tile output dir").Default("").Short('t').String()
	sr           = kingpin.Flag("sr", "spatial reference system").Default("4326").Short('s').String()
	id           = kingpin.Flag("id", "parcel object id key").Default("OBJECTID").Short('i').String()
	pin          = kingpin.Flag("pin", "parcel pin key").Default("PIN").Short('p').String()
	format       = kingpin.Flag("format", "format").Default("geojson").Short('f').String()
	wait         = kingpin.Flag("wait", "query wait (ms)").Default("500").Short('w').Int()
	update       = kingpin.Flag("update", "update existing files").Default("false").Short('u').Bool()
	verbose      = kingpin.Flag("verbose", "verbose").Default("false").Short('v').Bool()
	veryVerbose  = kingpin.Flag("vv", "very verbose").Default("false").Bool()
)

func main() {

	kingpin.Parse()

	var tiler *parcelier.Tiler

	opts := parcelier.Options{
		URL:              *url,
		OutputFilepath:   *outputFilepath,
		Update:           *update,
		TimeWait:         *wait,
		ParcelLimit:      1000,
		SpatialReference: *sr,
		Format:           *format,
		ParcelID:         *id,
		ParcelPIN:        *pin,
		SaveTilePath:     *saveTilePath,
		Verbose:          *verbose,
		VeryVerbose:      *veryVerbose,
	}

	if *boundaryFilepath != "" {
		boundaryData, err := parcelier.LoadFile(*boundaryFilepath)
		if err != nil {
			log.Fatal(err)
		}
		boundaryFeature := parcelier.GetFeature(boundaryData)
		tiler = parcelier.NewTiler(boundaryFeature.Geometry, *zoom)
		log.Printf("%d tiles at zoom %d\n", tiler.NumTiles, tiler.Zoom)
	} else if *extent != "" {
		log.Printf("todo: add support for download by single extent tile input (z/x/y)\n")
		return
	} else {
		log.Printf("no inputs to download\n")
		return
	}

	parcelier.Run(tiler, opts)

	log.Printf("done! %d tiles processed\n", len(tiler.Set))
}
