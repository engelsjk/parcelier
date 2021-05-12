package main

import (
	"fmt"

	"github.com/engelsjk/parcelier"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	boundaryFile = kingpin.Flag("boundary", "boundary file (e.g. county)").Default("").Short('b').String()
	extent       = kingpin.Flag("extent", "extent tile (z/x/y)").Default("").Short('e').String()
	url          = kingpin.Flag("url", "mapserver layer url").Default("").Short('u').String()
	//
	agent = kingpin.Flag("agent", "user agent").Default("parcelier").Short('a').String()
	zoom  = kingpin.Flag("zoom", "base zoom").Default("13").Short('z').Int()
	//
	parcelsPath = kingpin.Flag("parcels", "parcel output dir").Default(".").Short('p').String()
	tilesPath   = kingpin.Flag("tiles", "tile output dir").Default("").Short('t').String()
	//
	id = kingpin.Flag("id", "object id").Default("OBJECTID").String()
	//
	sr     = kingpin.Flag("sr", "spatial reference system").Default("4326").String()
	format = kingpin.Flag("format", "format").Default("geojson").Short('f').String()
	//
	parcelLimit = kingpin.Flag("limit", "object limit at zoom").Default("1000").Int()
	wait        = kingpin.Flag("wait", "query wait (ms)").Default("500").Int()
	update      = kingpin.Flag("update", "update existing files").Default("false").Bool()
	//
	verbose     = kingpin.Flag("verbose", "verbose").Default("false").Short('v').Bool()
	veryVerbose = kingpin.Flag("vv", "very verbose").Default("false").Bool()
	//
	infoOnly = kingpin.Flag("info", "tile info only").Default("false").Bool()
)

func main() {

	kingpin.Parse()

	var tiler *parcelier.Tiler

	opts := parcelier.Options{
		URL:              *url,
		Agent:            *agent,
		ParcelsPath:      *parcelsPath,
		Update:           *update,
		TimeWait:         *wait,
		ParcelLimit:      *parcelLimit,
		SpatialReference: *sr,
		Format:           *format,
		ParcelID:         *id,
		TilesPath:        *tilesPath,
		Verbose:          *verbose,
		VeryVerbose:      *veryVerbose,
	}

	if *extent != "" {
		fmt.Println("extent tile input not supported yet")
		return
	}

	if *boundaryFile == "" {
		fmt.Println("no inputs to download")
		return
	}

	if *url == "" {
		fmt.Println("no url provided")
		return
	}

	if *parcelsPath != "." {
		if !parcelier.DirExists(*parcelsPath) {
			fmt.Printf("parcels output path '%s' does not exist\n", *parcelsPath)
			return
		}
	}

	if *tilesPath != "" {
		if !parcelier.DirExists(*tilesPath) {
			fmt.Printf("tiles output path '%s' does not exist\n", *tilesPath)
			return
		}
	}

	//

	b, err := parcelier.LoadFile(*boundaryFile)
	if err != nil {
		fmt.Printf("unable to load boundary file %s\n", *boundaryFile)
		return
	}

	boundaryFeature, err := parcelier.GetFeature(b)
	if err != nil {
		fmt.Printf("unable to load boundary feature\n")
		return
	}

	tiler = parcelier.NewTiler(boundaryFeature.Geometry, *zoom)

	parcelier.PrintInfo(*boundaryFile, tiler)
	if *infoOnly {
		return
	}

	parcelier.Run(tiler, opts)
	parcelier.PrintDone(tiler)
	return
}
