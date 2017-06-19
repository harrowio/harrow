package projector

import (
	"fmt"
	"os"
	"testing"

	"github.com/boltdb/bolt"
)

func TestBoltDBIndex(t *testing.T) {
	filename := fmt.Sprintf("/tmp/bolt_db_index_test.db.%d", os.Getpid())
	os.Remove(filename)
	boltDB, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filename)
	NewIndexTest(func() Index {
		return NewBoltDBIndex(boltDB)
	}).Run(t)
}
