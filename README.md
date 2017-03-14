# Cesium Tile Server

## ekerner-com Fork of Cesium Terrain Server by geo-data

[cesium-terrain-server by the GeoData Institute](https://github.com/geo-data/cesium-terrain-server)

If youre already familiar with cesium-terrain-server
you can search for ekerner-com in this doc to find my added notes.

## Description

A basic server for serving up filesystem based tilesets representing
[Cesium.js](http://cesiumjs.org/) terrain models.  

ekerner-com added support for png raster tilesets, ssl/tls/https, and BasicAuth by htpasswd file with configurable HTTP Realm.

The resources served up are
intended for use with the
[`CesiumTerrainProvider`](http://cesiumjs.org/Cesium/Build/Documentation/CesiumTerrainProvider.html)
JavaScript class present in the Cesium.js client.

This has specifically been created for easing the development and testing of
terrain tilesets created using the
[Cesium Terrain Builder](https://github.com/geo-data/cesium-terrain-builder)
tools.

This project also provides a [Docker](https://www.docker.com/) container to
further simplify deployment of the server and testing of tilesets.  See the
[Docker Registry](https://registry.hub.docker.com/u/geodata/cesium-tile-server/)
for further details.

## Usage

The terrain server is a self contained binary with the following command line
options:

```sh
cesium-tile-server
  -base-terrain-url="/tilesets" # base url prefix under which all tilesets are served
  -cache-limit=1.00MB # the memory size in bytes beyond which resources are not cached. Other memory units can be specified by suffixing the number with kB, MB, GB or TB
  -dir="." # the root directory under which tileset directories reside
  -log-level=notice # level at which logging occurs. One of crit, err, notice, debug
  -memcached="" # (optional) memcached connection string for caching tiles e.g. localhost:11211
  -no-request-log=false # do not log client requests for resources
  -port=8000 # the port on which the server listens
  -web-dir="" # (optional) the root directory containing static files to be served
  -ssl-cert="" # (optional) server public key file. enables ssl. see notes at https://golang.org/pkg/net/http/#ListenAndServeTLS
  -ssl-key="" # (required if ssl-cert passed) server private key file
  -auth-file="" # (optional) htpasswd file. enables BasicAuth. doesnt apply to web-dir
  -auth-realm="Cesium Tile Server v1.0" # (optional) sets HTTP Realm for BasicAuth
```

Assume you have the following (small) terrain tileset (possibly created with
[`ctb-tile`](https://github.com/geo-data/cesium-terrain-builder#ctb-tile)):

```
/data/tilesets/terrain/srtm/
├── 0
│   └── 0
│       └── 0.terrain
├── 1
│   └── 1
│       └── 1.terrain
├── 2
│   └── 3
│       └── 3.terrain
└── 3
    └── 7
        └── 6.terrain
```

To serve this tileset on port `8080`, you would run the following command:

```sh
cesium-tile-server -dir /data/tilesets/terrain -port 8080
```

To serve a png raster tileset on port 8443 with ssl and basicauth:

```sh
cesium-tile-server -dir /data/tilesets/png -port 8443 -ssl-cert /etc/ssl/cert/cert.ca.bundle -ssl-key /etc/ssl/cert/server.key -auth-file /data/tilesets/.htpasswd
```

The tiles would then be available under <http://localhost:8080/tilesets/srtm/>
(e.g. <http://localhost:8080/tilesets/srtm/0/0/0.terrain> for the root tile or
<http://localhost:8080/tilesets/srtm/0/0/0.png>) for png raster tilesets.
This URL, for instance, is what you would use when configuring
[`CesiumTerrainProvider`](http://cesiumjs.org/Cesium/Build/Documentation/CesiumTerrainProvider.html)
in the Cesium client.

Serving up additional tilesets is simply a matter of adding the tileset as a
subdirectory to `/data/tilesets/terrain/`.  For example, adding a tileset
directory called `lidar` to that location will result in the tileset being
available under <http://localhost:8080/tilesets/lidar/>.

Note that the `-web-dir` option can be used to serve up static assets on the
filesystem in addition to tilesets.  This makes it easy to use the server to
prototype and develop web applications around the terrain data.

### `layer.json`

The `CesiumTerrainProvider` Cesium.js class requires that a `layer.json`
resource is present describing the terrain tileset.  The `ctb-tile` utility does
not create this file.  If a `layer.json` file is present in the root directory
of the tileset then this file will be returned by the server when the client
requests it.  If the file is not found then the server will return a default
resource.

ekerner-com: at the time of writing this you will need to manually make your png tileset layer.json files like:

```js
{
	"tilejson": "2.1.0",
	"name": "Multispec1-mercator-png",
	"version": "1.0.0",
	"description": "HTM Complete - Drone Mapping - Multispec1-mercator-png",
	"attribution": "<a href='http://www.htmcomplete.com.au/products/ebee-droning'>HTM Complete - Drone Mapping</a>",
	"scheme": "tms", // xyz|tms
	"tiles": ["{z}/{x}/{y}.png"], // terrain|png
	"bounds": [-180, -85.05112877980659, 180, 85.0511287798066],
	"center": [-76.275329586789, 39.153492567373, 8],
	"minzoom": 0, // 0-22
	"maxzoom": 22 // 0-22; >= minzoom
}
```

### Root tiles

The Cesium javascript client requires that the two top level tiles representing
zoom level `0` are always present.  These tiles are represented by the
`0/0/0.terrain` and `0/1/0.terrain` resources. When creating tilesets using the
`ctb-tile` utility only one of these tiles will be generated *unless* the source
terrain dataset intersects with the prime meridian.  The terrain server
addresses this issue by serving up a blank terrain tile if a top level tile is
requested which does not also exist on the filesystem.

### Caching tiles with Memcached

The terrain server can use a memcache server to cache tileset data. It is
important to note that the terrain server does not use the cache itself, it only
populates it for each request.  The idea is that a reverse proxy attached to the
memcache (such as Nginx) will first attempt to fulfil a request from the cache
before falling back to the terrain server, which will then update the cache.

Enabling this functionality requires specifying the network address of a
memcached server (including the port) using the `-memcached` option.  E.g. A
memcached server running at `memcache.me.org` on port `11211` can be used as
follows:

```sh
cesium-tile-server -dir /data/tilesets/terrain -memcached memcache.me.org:11211
```

If present, the terrain server uses the value of the custom `X-Memcache-Key`
header as the memcache key, otherwise it uses the value of the request URI.  A
minimal Nginx configuration setting `X-Memcache-Key` is as follows:

```
server {
    listen 80;

    server_name localhost;

    root /var/www/app;
    index index.html;

    location /tilesets/ {
        set            $memcached_key "tiles$request_uri";
        memcached_pass memcached:11211;
        error_page     404 502 504 = @fallback;
        add_header Access-Control-Allow-Origin "*";

        location ~* \.terrain$ {
            add_header Content-Encoding gzip;
        }
    }

    location @fallback {
        proxy_pass     http://tiles:8000;
        proxy_set_header X-Memcache-Key $memcached_key;
    }
}
```

The `-cache-limit` option can be used in conjunction with the above to change
the memory limit at which resources are considered to large for the cache.

## Installation

The server is written in [Go](http://golang.org/) and requires Go to be present
on the system when compiling it from source.  As such, it should run everywhere
that Go does.  Assuming that you have set the
[GOPATH](https://golang.org/cmd/go/#hdr-GOPATH_environment_variable),
installation is a matter of running `go install`:

```sh
go get github.com/ekerner-com/cesium-tile-server/cmd/cesium-tile-server
```

A program called `cesium-tile-server` should then be available under your
`GOPATH` (or `GOBIN` location if set).

## Additional notes by ekerner-com

### cesium-terrain-builder

ctb-tile util can generate raster png tilesets by adding the switch: 
`--output-format PNG`
Which is dependant on your gdal supporting PNG, see
```sh
gdalinfo --format png
```

### TLS/SSL - HTTPS

If youre using a cert authority then your ssl-key file should contain your server public key, followed by your certificate authority bundle. like
```sh
cd /etc/ssl/cert/ # wherevs your cert files are
cat server.pub.key server.pub.ca.bundle.cert > server.pub.certchain.cert
```
As per usual the certificate within must be in order of heirarchy from server to certficate auth peer(s).

### htpasswd

This is great for us, we can do multilevel authentication by
1. app/site logs into api service by user input, gets login token,
2. app/site uses token to acquire a `temp tile server auth` from the same api service,
3. app/site uses said auth to make request to the tile server.
4. app calls to api service to delete the tmp user when done ..
  * and/or server deletes tmp users after certain time

And in step 2 htpasswd makes it very easy, because the api can:
```sh
# add user tmpusr-time-crc32
htpasswd -b .htpasswd tmpusr-1489497656-473d4a25 db46d12a
# delete user (well, their login only hopefully)
htpasswd -D .htpasswd tmpusr-1489497656-473d4a25 # I think
# see of course for extras
man htpasswd
```

You can bulk delete from the .htpasswd file (like with grep/cat | sed/perl), removing users which 
start with tmpusr- and 
time - timestamp < expiry

### Docker

I didnt touch the docker stuff except renames and a note in the README.md. So to use the ssl and/or basicauth in docker someone would need to do something like:
1. pass the `-ssl-cert`, `-ssl-key`, and `-auth-file` args thru to docker
2. docker forward args to cesium-tile-server
3. bind docker and cesium-tile-server ssl ports

## Credits - added by ekerner-com

### motivation

This update was made for and funded by the HTM Complete (TM) Australia Drone Mapping and Crop Analysis Project - a part of the Hydrotech Monitoring Pty Ltd Australia ~ HTM Complete (TM) Agriculural Management Technologies Development Project 2015-2017.
[HTM Complete - Drone Mapping](http://www.htmcomplete.com.au/products/ebee-droning)

### libs

ekerner-com: Aside from employing
[cesium-terrain-server by the GeoData Institute](https://github.com/geo-data/cesium-terrain-server)
I obtained the htpasswd logic from
[go-htpasswd by tg123](https://github.com/tg123/go-htpasswd)

### Authors and contributors (that I know of)

**GeoData Institute**
[GeoData Institute - cesium-terrain-server and cesium-terrain-builder](https://github.com/geo-data/)

**Homme Zwaagstra**
Homme Zwaagstra <hrz@geodata.soton.ac.uk>

**HTM Complete**
[HTM Complete Australia](http://www.htmcomplete.com.au/)

**Boshi Lain**
[Boshi Lain - htpasswd](https://github.com/tg123)

**eKerner.com**
[eKerner Pty Ltd Australia - cesium-tile-server](http://www.ekerner.com/)

**Eugene Kerner**
Eugene Kerner <ekerner@ekerner.com>

## Developing

The code has been developed on a Linux platform. After downloading the package
you should be able to run `make` from the project root to build the server,
which will be available as `./bin/cesium-tile-server`.

ekerner-com: I added rule `installnoget` tp the Makefile for building without fetching - so I could dev on it.

Executing `make docker-local` will create a docker image tagged
`geodata/cesium-tile-server:local` which when run with a bind mount to the
project source directory is very handy for developing and testing.

## Issues and Contributing

Please report bugs or issues using the
[GitHub issue tracker](https://github.com/ekerner-com/cesium-tile-server).

Code and documentation contributions are very welcome, either as GitHub pull
requests or patches.
ekerner-com: OK I didnt read the above before I forked. @geo-data if you prefer we can merge my changes back into your repo and delete this one?

ekerner.com: My 1st go project to be honest, love the lang so far.
If there are things which arent Ideal they are
1. someone could make a super/base class for 
`handers/terrain.go`
`handers/pngtile.go`
and perhaps
`handers/terrain.go`
as there is duplicate code in each
2. would be nice to add support for gzip and/or deflate: 
you could say zgip all of the images if you had a poory compressed png tilesets where they sit. Just needs different headers and perhaps diff store data type.
3. the handlers/layer.json default file extension is .terrain: nice if you could pass in the default ext eg:
`-default-format png // terrain|png`

## License

The [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0).

## Contact

Homme Zwaagstra <hrz@geodata.soton.ac.uk>

ekerner-com: Eugene Kerner <ekerner@ekerner.com>
