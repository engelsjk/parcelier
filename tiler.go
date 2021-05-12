package parcelier

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/paulmach/orb"
	geojson "github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/paulmach/orb/maptile/tilecover"
)

type Tiler struct {
	Zoom      int
	Set       maptile.Set
	NumTiles  int
	TileOrder []string
}

func NewTiler(geom orb.Geometry, zoom int) *Tiler {
	tileSet := NewTileset(geom, zoom)
	numTiles := len(tileSet)
	tileOrder := []string{}

	for tile := range tileSet {
		t := (&Tile{tile}).GetTileString()
		tileOrder = append(tileOrder, t)
	}
	sort.Strings(tileOrder)

	return &Tiler{
		Zoom:      zoom,
		Set:       tileSet,
		NumTiles:  numTiles,
		TileOrder: tileOrder,
	}
}

func NewTileset(geom orb.Geometry, zoom int) maptile.Set {
	return tilecover.Geometry(geom, maptile.Zoom(zoom))
}

func (t *Tiler) GetTileAtIndex(idx int) (Tile, error) {
	ts := t.TileOrder[idx]
	return ConvertStringToTile(ts)
}

func (t *Tiler) GetGeoJSON() ([]byte, error) {
	fc := geojson.NewFeatureCollection()
	fc.Features = make([]*geojson.Feature, 0, len(t.Set))
	for maptile := range t.Set {
		tile := Tile{maptile}
		f := geojson.NewFeature(tile.Bound().ToPolygon())
		f.Properties["tile"] = tile.GetTileString()
		f.Properties["extent"] = tile.GetExtentString()
		fc.Append(f)
	}
	rawJSON, err := json.MarshalIndent(fc, "", " ")
	if err != nil {
		return nil, err
	}
	return rawJSON, nil
}

type Tile struct {
	maptile.Tile
}

func (t *Tile) GetTileString() string {
	return fmt.Sprintf("%d/%d/%d", t.Z, t.X, t.Y)
}

func (t *Tile) GetExtentString() string {
	b := t.Bound()
	return fmt.Sprintf("%f,%f,%f,%f", b.Left(), b.Bottom(), b.Right(), b.Top())
}

func (t *Tile) GetTilePath(path string) string {
	filename := fmt.Sprintf("tile_%d_%d_%d.geojson", t.Z, t.X, t.Y)
	filePath := filepath.Join(path, filename)
	return filePath
}

func (t *Tile) GetParcelPath(path string) string {
	filename := fmt.Sprintf("parcels_%d_%d_%d.geojson", t.Z, t.X, t.Y)
	filePath := filepath.Join(path, filename)
	return filePath
}

func ConvertStringToTile(t string) (Tile, error) {
	zxy := strings.Split(t, "/")
	if len(zxy) != 3 {
		return Tile{}, fmt.Errorf("tile string must be {z}/{x}/{y}")
	}

	z, err := strconv.ParseUint(zxy[0], 10, 32)
	if err != nil {
		return Tile{}, fmt.Errorf("z must be numeric")
	}
	x, err := strconv.ParseUint(zxy[1], 10, 32)
	if err != nil {
		return Tile{}, fmt.Errorf("x must be numeric")
	}
	y, err := strconv.ParseUint(zxy[2], 10, 32)
	if err != nil {
		return Tile{}, fmt.Errorf("y must be numeric")
	}

	return Tile{
		maptile.Tile{X: uint32(x), Y: uint32(y), Z: maptile.Zoom(uint32(z))},
	}, nil
}
