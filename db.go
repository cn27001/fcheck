package fcheck

import (
	"log"

	"github.com/boltdb/bolt"
)

//DB represents the underlying datastore that stores the actual filesystem entries
type DB struct {
	dbfile string
	bucket []byte
	db     *bolt.DB
}

//NewDB returns new instance of DB
func NewDB(dbfname string) *DB {
	return &DB{dbfile: dbfname,
		bucket: []byte("fs")}
}

//Start performs any needed initialization
func (r *DB) Start() error {
	if r.db == nil {
		//make db
		db, err := bolt.Open(r.dbfile, 0600, nil)
		if err != nil {
			log.Printf("Trouble in bolt.Open: %s\n", err.Error())
			return err
		}
		r.db = db
		//make bucket
		if err := r.db.Update(func(tx *bolt.Tx) error {
			if _, err := tx.CreateBucketIfNotExists(r.bucket); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

//Stop performs any needed cleanup
func (r *DB) Stop() error {
	if r.db != nil {
		if err := r.db.Close(); err != nil {
			return err
		}
	}
	return nil
}

//Set puts an entry in the datastore
func (r *DB) Set(key, val []byte) error {
	err := r.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(r.bucket)
		return bucket.Put(key, val)
	})
	return err
}

//Get retreives an entry from the datastore
func (r *DB) Get(key []byte) ([]byte, error) {
	var val []byte
	err := r.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(r.bucket)
		val = bucket.Get(key)
		return nil
	})
	return val, err
}

//Map iterates over all entries in datastore and calls function f on them
func (r *DB) Map(f DBMapFunc) error {
	err := r.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(r.bucket)
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			f(k, v)
		}
		return nil
	})
	return err
}

//DBMapFunc is the callback function definition used by Map
type DBMapFunc func(key, value []byte)
