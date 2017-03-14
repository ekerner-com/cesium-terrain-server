package handlers

import (
	"errors"
	"fmt"
	"github.com/ekerner-com/cesium-tile-server/htpasswd"
	"github.com/ekerner-com/cesium-tile-server/assets"
	"github.com/ekerner-com/cesium-tile-server/log"
	"github.com/ekerner-com/cesium-tile-server/stores"
	"github.com/gorilla/mux"
	"net/http"
)

// An HTTP handler which returns a png tile resource
func PngtileHandler(store stores.Storer, authFile string, authRealm string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			t   stores.Pngtile
			err error
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

		// get the tile coordinate from the URL
		err = t.ParseCoord(vars["x"], vars["y"], vars["z"])
		if err != nil {
			return
		}

		// Try and get a tile from the store
		err = store.TilePng(vars["tileset"], &t)
		if err == stores.ErrNoItem {
			if store.TilesetStatus(vars["tileset"]) == stores.NOT_FOUND {
				err = nil
				http.Error(w,
					fmt.Errorf("The tileset `%s` does not exist", vars["tileset"]).Error(),
					http.StatusNotFound)
				return
			}

			if t.IsRoot() {
				// serve up a blank tile as it is a missing root tile
				data, err := assets.Asset("data/blank.png")
				if err != nil {
					return
				} else {
					err = t.PutImage(data)
					if err != nil {
						return
					}
				}
			} else {
				err = nil
				http.Error(w, errors.New("The png tile does not exist").Error(), http.StatusNotFound)
				return
			}
		} else if err != nil {
			return
		}

		body, err := t.GetImage()
		if err != nil {
			return
		}

		// send the tile to the client
		log.Debug("PngtileHandler sending /"+vars["tileset"]+"/"+vars["z"]+"/"+vars["x"]+"/"+vars["y"]+".png")
		headers := w.Header()
		headers.Set("Content-Type", "image/png")
		//headers.Set("Content-Encoding", "gzip")
		w.Write(body)
	}
}
