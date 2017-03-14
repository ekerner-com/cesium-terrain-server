package stores

import (
	"strconv"
)

// Representation of a terrain tile. This includes the x, y, z coordinate and
// the byte sequence of the tile itself. Note that terrain tiles are gzipped.
type Pngtile struct {
	value   []byte
	X, Y, Z uint64
}

func (this *Pngtile) GetImage() ([]byte, error) {
	return this.value, nil
}

func (this *Pngtile) PutImage(data []byte) error {
	this.value = data
	return nil
}

// IsRoot returns true if the tile represents a root tile.
func (self *Pngtile) IsRoot() bool {
	return self.Z == 0 &&
		(self.X == 0 || self.X == 1) &&
		self.Y == 0
}

// Parse x, y, z string coordinates and assign them to the Pngtile
func (self *Pngtile) ParseCoord(x, y, z string) error {
	xi, err := strconv.ParseUint(x, 10, 64)
	if err != nil {
		return err
	}

	yi, err := strconv.ParseUint(y, 10, 64)
	if err != nil {
		return err
	}

	zi, err := strconv.ParseUint(z, 10, 64)
	if err != nil {
		return err
	}

	self.X = xi
	self.Y = yi
	self.Z = zi

	return nil
}
