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
		if _, err := tx.CreateBucketIfNotExists(boltLinksBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(boltUsersBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(boltAdminsBucket); err != nil {
			return err
		}
		return nil
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

// GetLink returns a Link object for a given key.
func (d *DB) GetLink(key string) (Link, bool, error) {
	var link Link
	var exists bool
	if err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(boltLinksBucket)
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

// PutLink saves a Link object under a given key.
// If the link already exists, `user` must match the link's owner, unless `force` is true.
func (d *DB) PutLink(key, value, user string, force bool) error {
	v, err := Link{Key: key, Value: value, Owner: user}.Bytes()
	if err != nil {
		return err
	}
	return d.db.Update(func(tx *bolt.Tx) error {
		k := []byte(key)
		b := tx.Bucket(boltLinksBucket)

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

// DeleteLink clears a given key.
// `user` must match the link's owner, unless `force` is true.
func (d *DB) DeleteLink(key, user string, force bool) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		k := []byte(key)
		b := tx.Bucket(boltLinksBucket)

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

// ListLinks returns links belonging to a user.
// If `all` is set, all links (including those owned by other users) are returned.
func (d *DB) ListLinks(user string, all bool) ([]Link, error) {
	var links []Link
	if err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(boltLinksBucket)
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
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return links, nil
}

// AddUser adds a given user name into the DB.
func (d *DB) AddUser(user string) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		// From https://github.com/boltdb/bolt/blob/master/README.md:
		// [...] a zero-length value set to a key [...] is different than the key not existing.
		return tx.Bucket(boltUsersBucket).Put([]byte(user), []byte{})
	})
}

// IsUser returns a bool indicating whether the given user name is present in DB.
func (d *DB) IsUser(user string) (bool, error) {
	v := false
	if err := d.db.View(func(tx *bolt.Tx) error {
		v = !(tx.Bucket(boltUsersBucket).Get([]byte(user)) == nil)
		return nil
	}); err != nil {
		return false, err
	}
	return v, nil
}

// ListUsers returns the list of user names stored in DB.
func (d *DB) ListUsers() ([]string, error) {
	var users []string
	if err := d.db.View(func(tx *bolt.Tx) error {
		if err := tx.Bucket(boltUsersBucket).ForEach(func(k, v []byte) error {
			users = append(users, string(k))
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return users, nil
}

// DeleteUser deletes a given user (and optionally their Links) from the DB.
func (d *DB) DeleteUser(user string, keepLinks bool) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		if err := tx.Bucket(boltUsersBucket).Delete([]byte(user)); err != nil {
			return err
		}

		if keepLinks {
			return nil
		}

		b := tx.Bucket(boltLinksBucket)
		if err := b.ForEach(func(k, v []byte) error {
			l, err := LinkFromBytes(v)
			if err != nil {
				log.Printf("malformed link under key %q", string(k))
				return err
			}
			if l.Owner == user {
				if err := b.Delete(k); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
}

// IsAdmin returns a bool indicating whether the given admin name is present in DB.
func (d *DB) IsAdmin(user string) (bool, error) {
	v := false
	if err := d.db.View(func(tx *bolt.Tx) error {
		v = !(tx.Bucket(boltAdminsBucket).Get([]byte(user)) == nil)
		return nil
	}); err != nil {
		return false, err
	}
	return v, nil
}

// UpdateAdmins replaces the list of admin names stored in DB.
func (d *DB) UpdateAdmins(admins []string) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket(boltAdminsBucket); err != nil && err != bolt.ErrBucketNotFound {
			return err
		}

		b, err := tx.CreateBucketIfNotExists(boltAdminsBucket)
		if err != nil {
			return err
		}

		for _, v := range admins {
			if err := b.Put([]byte(v), []byte{}); err != nil {
				return err
			}
		}
		return nil
	})
}

// DB is a wrapper on top of BoltDB with methods for shortlink reading and manipulation.
type DB struct {
	db *bolt.DB
}

var boltLinksBucket = []byte("LINKS")
var boltUsersBucket = []byte("USERS")
var boltAdminsBucket = []byte("ADMINS")
