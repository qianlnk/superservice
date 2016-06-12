package storage

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
)

var (
	ErrNoKeyExists = errors.New("no key exists")
)

type KeyValue struct {
	Key   string
	Value []byte
}

type Bolt struct {
	db     *bolt.DB
	bucket []byte
}

var boltDB *bolt.DB

func init() {
	var err error
	boltDB, err = bolt.Open("superservice.db", 0600, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func NewBolt(db *bolt.DB, bucket string) *Bolt {
	return &Bolt{
		db:     db,
		bucket: []byte(bucket),
	}
}

func GetBolt(bucket string) *Bolt {
	return &Bolt{
		db:     boltDB,
		bucket: []byte(bucket),
	}
}

func (b *Bolt) Put(key string, value interface{}) error {
	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(b.bucket)
		if err != nil {
			return err
		}
		err = bucket.Put([]byte(key), buf)
		if err != nil {
			return err
		}
		return nil
	})
}

func (b *Bolt) Update(key string, value interface{}) error {
	err := b.Delete(key)
	if err != nil {
		return err
	}
	return b.Put(key, value)
}

func (b *Bolt) Get(key string) (*KeyValue, error) {
	var value []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.bucket)
		if bucket == nil {
			return ErrNoKeyExists
		}

		val := bucket.Get([]byte(key))
		if val == nil {
			return ErrNoKeyExists
		}
		value = make([]byte, len(val))
		copy(value, val)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &KeyValue{
		Key:   key,
		Value: value,
	}, nil
}

func (b *Bolt) Delete(key string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.bucket)
		if bucket == nil {
			return nil
		}
		return bucket.Delete([]byte(key))
	})
}

func (b *Bolt) Exists(key string) (bool, error) {
	var exists bool
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.bucket)
		if bucket == nil {
			return nil
		}

		val := bucket.Get([]byte(key))
		exists = val != nil
		return nil
	})
	return exists, err
}

func (b *Bolt) List() (kvs []*KeyValue, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.bucket)
		if bucket == nil {
			return nil
		}
		bucket.ForEach(func(key, v []byte) error {
			value := make([]byte, len(v))
			copy(value, v)

			kvs = append(kvs, &KeyValue{
				Key:   string(key),
				Value: value,
			})
			return nil
		})
		return nil
	})
	return kvs, err
}

func (b *Bolt) Close() {
	b.db.Close()
}
