package database

import (
	"bytes"
	"encoding/gob"
)

// Link contains shortlink metadata.
type Link struct {
	Key   string
	Value string
	Owner string
}

// Bytes returns the serialized representation of Link.
func (l Link) Bytes() ([]byte, error) {
	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(l); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// LinkFromBytes creates a Link from a serialized representation.
func LinkFromBytes(b []byte) (Link, error) {
	var l Link
	if err := gob.NewDecoder(bytes.NewReader(b)).Decode(&l); err != nil {
		return Link{}, err
	}
	return l, nil
}
