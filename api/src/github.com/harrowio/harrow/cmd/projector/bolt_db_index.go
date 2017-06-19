package projector

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
)

type BoltDBIndex struct {
	db         *bolt.DB
	bucketName string
}

func NewBoltDBIndex(db *bolt.DB) *BoltDBIndex {
	return &BoltDBIndex{
		db:         db,
		bucketName: "projector",
	}
}

func (self *BoltDBIndex) Update(do func(tx IndexTransaction) error) error {
	return self.db.Update(func(boltTx *bolt.Tx) error {
		return do(NewBoltDBIndexTransaction(boltTx, self.bucketName))
	})
}

type BoltDBIndexTransaction struct {
	tx         *bolt.Tx
	bucketName string
}

func NewBoltDBIndexTransaction(tx *bolt.Tx, bucketName string) *BoltDBIndexTransaction {
	return &BoltDBIndexTransaction{
		tx:         tx,
		bucketName: bucketName,
	}
}

func (self *BoltDBIndexTransaction) Get(uuid string, dest interface{}) error {
	bucket := self.tx.Bucket([]byte(self.bucketName))
	if bucket == nil {
		return fmt.Errorf("%T code=bucket_not_found key=%s", self, uuid)
	}

	asJSON := bucket.Get([]byte(uuid))
	if asJSON == nil {
		return fmt.Errorf("%T code=not_found key=%s", self, uuid)
	}

	return json.Unmarshal(asJSON, dest)
}

func (self *BoltDBIndexTransaction) Put(uuid string, src interface{}) error {
	bucket, err := self.tx.CreateBucketIfNotExists([]byte(self.bucketName))
	if err != nil {
		return err
	}

	asJSON, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return bucket.Put([]byte(uuid), asJSON)
}
