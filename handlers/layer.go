package handlers

import (
	"fmt"
	"github.com/ekerner-com/cesium-tile-server/htpasswd"
	"github.com/ekerner-com/cesium-tile-server/log"
	"github.com/ekerner-com/cesium-tile-server/stores"
	"github.com/gorilla/mux"
	"net/http"
)

// An HTTP handler which returns a tileset's `layer.json` file
func LayerHandler(store stores.Storer, authFile string, authRealm string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err   error
			layer []byte
		)

		defer func() {
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Err(err.Error())
			}
		}()

		vars := mux.Vars(r)

		// Authenticate
		if authFile != "" {
			log.Debug(fmt.Sprintf("Authenticating using auth-file: %s", authFile))
			basicAuth, err := htpasswd.New(authRealm, authFile, htpasswd.DefaultSystems, nil)
			if err != nil {
				http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
				log.Err(err.Error())
				return
			}
			if !basicAuth.Authenticate(w, r) {
				log.Notice("Auth failed for "+r.RemoteAddr)
				return
			}
		}

		// Try and get a `layer.json` from the stores
		layer, err = store.Layer(vars["tileset"])
		if err == stores.ErrNoItem {
			err = nil // don't persist this error
			if store.TilesetStatus(vars["tileset"]) == stores.NOT_FOUND {
				http.Error(w,
					fmt.Errorf("The tileset `%s` does not exist!", vars["tileset"]).Error(),
					http.StatusNotFound)
				return
			}

			// the directory exists: send the default `layer.json`
			layer = []byte(`{
  "tilejson": "2.1.0",
  // "format": "heightmap-1.0", // not a TileJSON 2.1.0 property
  "version": "1.0.0",
  "scheme": "tms",
  // TODO: set .terrain to .png for raster tilesets,
  //  not sure how tho, cant reliably test for 0/0/0.png
  //  so this is broken in raster sets, in which case
  //  need to place file manually, see README.md
  "tiles": ["{z}/{x}/{y}.terrain"]
}`)
		} else if err != nil {
			return
		}

		headers := w.Header()
		headers.Set("Content-Type", "application/json; profile=/layer.schema.json#") // place layer.schema.json in web-dir
		w.Write(layer)
	}
}
