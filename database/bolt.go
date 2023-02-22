package database

import (
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

// NewBolt creates a DB object using the provided backing file.
func NewBolt(file string) (*DB, error) {
	db, err := bolt.Open(file, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(boltLinksBucket))
		return err
	}); err != nil {
		db.Close()
		return nil, err
	}
	return &DB{db: db}, nil
}

// Close closes the DB.
func (d *DB) Close() error {
	return d.db.Close()
}

// Get returns a Link object for a given key.
func (d *DB) Get(key string) (Link, bool, error) {
	var link Link
	var exists bool
	if err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(boltLinksBucket))
		bytes := b.Get([]byte(key))
		if len(bytes) == 0 {
			exists = false
			return nil
		}
		l, err := LinkFromBytes(bytes)
		if err != nil {
			log.Printf("malformed link under key %q", key)
			return err
		}
		link = l
		exists = true
		return nil
	}); err != nil {
		return Link{}, false, err
	}
	return link, exists, nil
}

// Put saves a Link object under a given key.
// If the link already exists, `user` must match the link's owner, unless `force` is true.
func (d *DB) Put(key, value, user string, force bool) error {
	v, err := Link{Key: key, Value: value, Owner: user}.Bytes()
	if err != nil {
		return err
	}
	return d.db.Update(func(tx *bolt.Tx) error {
		k := []byte(key)
		b := tx.Bucket([]byte(boltLinksBucket))

		bytes := b.Get(k)
		if len(bytes) == 0 {
			return b.Put(k, v)
		}

		l, err := LinkFromBytes(bytes)
		if err != nil {
			log.Printf("malformed link under key %q", key)
			return err
		}
		if force || l.Owner == user {
			return b.Put(k, v)
		}
		return fmt.Errorf("link owned by someone else")
	})
}

// Delete clears a given key.
// `user` must match the link's owner, unless `force` is true.
func (d *DB) Delete(key, user string, force bool) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		k := []byte(key)
		b := tx.Bucket([]byte(boltLinksBucket))

		bytes := b.Get(k)
		if len(bytes) == 0 {
			return nil
		}

		l, err := LinkFromBytes(bytes)
		if err != nil {
			log.Printf("malformed link under key %q", key)
			return err
		}
		if force || l.Owner == user {
			return b.Delete(k)
		}
		return fmt.Errorf("link owned by someone else")
	})
}

// List returns links belonging to a user.
// If `all` is set, all links (including those owned by other users) are returned.
func (d *DB) List(user string, all bool) ([]Link, error) {
	var links []Link
	if err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(boltLinksBucket))
		if err := b.ForEach(func(k, v []byte) error {
			l, err := LinkFromBytes(v)
			if err != nil {
				log.Printf("malformed link under key %q", string(k))
				return err
			}
			if l.Owner != user && !all {
				return nil
			}
			links = append(links, l)
			return nil
		}); err != nil {
			log.Printf("listing failed: %v", err)
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return links, nil
}

// DB is a wrapper on top of BoltDB with methods for shortlink reading and manipulation.
type DB struct {
	db *bolt.DB
}

var boltLinksBucket = "LINKS"
