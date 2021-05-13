# parcelier

```parceler``` is a Go CLI that extracts and converts GeoJSON from an ESRI MapServer endpoint, mostly aimed at county parcel websites. In some regards, it's a more limited version of [openaddresses/pyesridump](https://github.com/openaddresses/pyesridump).

I've found it useful to pair with some other GeoJSON tools like [engelsjk/gjfunks](https://github.com/engelsjk/gjfunks) and [stedolan/jq](https://github.com/stedolan/jq).

## What

```parcelier``` is primarily intended for extracting parcel boundary data from US county GIS websites. At a high level, this tool takes in a boundary GeoJSON file of a specific county, tiles the boundary at a specified zoom level and then iteratively requests features from an ESRI MapServer REST endpoint for each tile boundary. The response data is returned as GeoJSON (or converted to such) and each tile worth of data is then saved as a GeoJSON file to disk.

For tile requests that exceeded the server feature limit (typically 1000 features), ```parcelier``` will up-tile to higher zoom levels until the requested number of features is below the limit to ensure complete coverage. This up-tiling process might look something like this.

![](images/tiles.png)

There are arguably better or more efficient or more reasonable ways to accomplish this specific task. This is just one way that I've found to work reasonably well.

## Help

```bash
parcelier --help
```

```bash
usage: parcelier [<flags>]

Flags:
      --help               Show context-sensitive help (also try --help-long and --help-man).
  -b, --boundary=""        boundary file (e.g. county)
  -e, --extent=""          extent tile (z/x/y)
  -u, --url=""             mapserver layer url
  -a, --agent="parcelier"  user agent
  -z, --zoom=13            base zoom
  -p, --parcels="."        parcel output dir
  -t, --tiles=""           tile output dir
      --id="OBJECTID"      object id
      --sr="4326"          spatial reference system
  -f, --format="geojson"   format
      --limit=1000         object limit at zoom
      --wait=500           query wait (ms)
      --update             update existing files
  -v, --verbose            verbose
      --vv                 very verbose
      --info               tile info only
```

## Commands

```bash
parcelier \
-b "boundary.geojson" \
--url "https://gis.website.com/arcgis/rest/services/Parcels/MapServer/0" \
-i "FID" \
-p "PIN" \
-f "geoJSON" \
-w 100 \
-z 13 \
-s "4326" \
-t "tiles" \
-o "parcels"
```

```bash
gjbuild \
-o "parcels.ndjson" \
--filter-key="PIN" \
--ndjson \
parcels
```

```bash
jq \
-c \
'.properties' \
parcels.ndjson > \
properties.ndjson
```
