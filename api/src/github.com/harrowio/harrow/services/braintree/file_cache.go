package braintree

import (
	"encoding/json"
	"os"

	braintreeAPI "github.com/lionelbarrow/braintree-go"
)

type FileCache struct {
	filename string
}

func NewFileCache(filename string) *FileCache {
	return &FileCache{
		filename: filename,
	}
}

func (self *FileCache) Clear() {
	os.Remove(self.filename)
}

func (self *FileCache) Entries() ([]*braintreeAPI.Plan, error) {
	result := []*braintreeAPI.Plan{}
	in, err := os.Open(self.filename)
	if err != nil {
		return nil, ErrCacheMiss
	}

	dec := json.NewDecoder(in)
	if err := dec.Decode(&result); err != nil {
		return nil, ErrCacheMiss
	}

	return result, nil
}

func (self *FileCache) Set(entries []*braintreeAPI.Plan) error {
	out, err := os.Create(self.filename)
	if err != nil {
		return ErrCacheInternal
	}
	defer out.Close()

	enc := json.NewEncoder(out)
	if err := enc.Encode(entries); err != nil {
		return ErrCacheInternal
	}

	return nil
}
