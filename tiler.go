package parcelier

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/paulmach/orb"
	geojson "github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/paulmach/orb/maptile/tilecover"
)

type Tiler struct {
	Zoom     int
	Set      maptile.Set
	NumTiles int
}

func NewTiler(geom orb.Geometry, zoom int) *Tiler {
	tileset := GetTileset(geom, zoom)
	return &Tiler{
		Zoom:     zoom,
		Set:      tileset,
		NumTiles: len(tileset),
	}
}

func GetTileset(geom orb.Geometry, zoom int) maptile.Set {
	return tilecover.Geometry(geom, maptile.Zoom(zoom))
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

func (t *Tile) GetFilepath(path string) string {
	filename := fmt.Sprintf("tile-z%d-x%d-y%d.geojson", t.Z, t.X, t.Y)
	filePath := filepath.Join(path, filename)
	return filePath
}
