package core

import (
	"errors"

	"github.com/speps/go-hashids"
)

// DecodeHashID decodes hashID to int
func DecodeHashID(salt, hashID string) (int, error) {
	hd := hashids.NewData()
	hd.MinLength = 8
	hd.Salt = salt
	ids, err := hashids.NewWithData(hd).DecodeWithError(hashID)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, errors.New("No id for hash")
	}
	return ids[0], nil
}
