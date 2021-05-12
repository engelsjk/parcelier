package main

import (
	"fmt"
	"log"
	"path"

	"github.com/engelsjk/parcelier"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	boundaryFile = kingpin.Flag("boundary", "boundary file").Default("").Short('b').String()
	extent       = kingpin.Flag("extent", "extent tile (z/x/y)").Default("").Short('e').String()
	url          = kingpin.Flag("url", "esri url").Default("").Short('u').String()
	//
	agent = kingpin.Flag("agent", "user agent").Default("parcelier").Short('a').String()
	zoom  = kingpin.Flag("zoom", "initial zoom").Default("13").Short('z').Int()
	//
	parcelsPath = kingpin.Flag("parcels", "parcel output dir").Default(".").Short('p').String()
	tilesPath   = kingpin.Flag("tiles", "tile output dir").Default("").Short('t').String()
	//
	id  = kingpin.Flag("id", "parcel object id key").Default("OBJECTID").String()
	pin = kingpin.Flag("pin", "parcel pin key").Default("PIN").String()
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
	info = kingpin.Flag("info", "tile info").Default("false").Bool()
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
		ParcelPIN:        *pin,
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

	log.Println(*url)

	if *url == "" {
		fmt.Println("no url provided")
		return
	}

	//

	boundaryData, err := parcelier.LoadFile(*boundaryFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	boundaryFeature := parcelier.GetFeature(boundaryData)

	tiler = parcelier.NewTiler(boundaryFeature.Geometry, *zoom)

	fmt.Printf("running boundary %s...\n", path.Base(*boundaryFile))
	fmt.Printf("%d tiles at zoom %d\n", tiler.NumTiles, tiler.Zoom)

	if *info {
		return
	}

	parcelier.Run(tiler, opts)

	log.Printf("done! %d tiles processed\n", len(tiler.Set))
	return

}
