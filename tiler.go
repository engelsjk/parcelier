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

// Tiler ...
type Tiler struct {
	Zoom      int
	Set       maptile.Set
	NumTiles  int
	TileOrder []string
}

// NewTiler ...
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

// NewTileset ...
func NewTileset(geom orb.Geometry, zoom int) maptile.Set {
	return tilecover.Geometry(geom, maptile.Zoom(zoom))
}

// GetTileAtIndex returns a Tile from the ordered Tile set at a given index.
func (t *Tiler) GetTileAtIndex(idx int) Tile {
	ts := t.TileOrder[idx]
	return ConvertStringToTile(ts)
}

// GetGeoJSON returns a GeoJSON Feature Collection for the Tiler set.
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

// Tile ...
type Tile struct {
	maptile.Tile
}

// GetTileString returns a 'z/x/y' tile string
func (t *Tile) GetTileString() string {
	return fmt.Sprintf("%d/%d/%d", t.Z, t.X, t.Y)
}

// GetExtentString returns a 'left,bottom,right,top' extent string
func (t *Tile) GetExtentString() string {
	b := t.Bound()
	return fmt.Sprintf("%f,%f,%f,%f", b.Left(), b.Bottom(), b.Right(), b.Top())
}

// GetTilePath returns a tile filepath string for the tile.
func (t *Tile) GetTilePath(path string) string {
	filename := fmt.Sprintf("tile_%d_%d_%d.geojson", t.Z, t.X, t.Y)
	filePath := filepath.Join(path, filename)
	return filePath
}

// GetParcelPath returns a parcel filepath string for the tile.
func (t *Tile) GetParcelPath(path string) string {
	filename := fmt.Sprintf("parcels_%d_%d_%d.geojson", t.Z, t.X, t.Y)
	filePath := filepath.Join(path, filename)
	return filePath
}

// ConvertStringToTile ...
func ConvertStringToTile(t string) Tile {
	zxy := strings.Split(t, "/")
	if len(zxy) != 3 {
		// fmt.Errorf("tile string must be {z}/{x}/{y}")
	}

	z, _ := strconv.ParseUint(zxy[0], 10, 32)
	// fmt.Errorf("z must be numeric")

	x, _ := strconv.ParseUint(zxy[1], 10, 32)
	// fmt.Errorf("x must be numeric")

	y, _ := strconv.ParseUint(zxy[2], 10, 32)
	// fmt.Errorf("y must be numeric")

	return Tile{
		maptile.Tile{X: uint32(x), Y: uint32(y), Z: maptile.Zoom(uint32(z))},
	}
}
