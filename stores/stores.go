package stores

import (
	"errors"
)

type TilesetStatus byte

const (
	NOT_SUPPORTED TilesetStatus = iota
	NOT_FOUND
	FOUND
)

var ErrNoItem = errors.New("item not found")

type Storer interface {
	TileTerrain(tileset string, tile *Terrain) error
	TilePng(tileset string, tile *Pngtile) error
	Layer(tileset string) ([]byte, error)
	TilesetStatus(tileset string) (status TilesetStatus)
}
