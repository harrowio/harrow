package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Dictionary is a generic mapping for seamlessly storing JSON objects
// in the database.
type Dictionary map[string]interface{}

func NewDictionary() Dictionary {
	return Dictionary{}
}

// Set associates key with value in the dictionary.
func (dictionary Dictionary) Set(key string, value interface{}) Dictionary {
	dictionary[key] = value
	return dictionary
}

// Get returns the value associated with key in the dictionary or nil
// if no such value exists.
func (dictionary Dictionary) Get(key string) interface{} {
	return dictionary[key]
}

func (dictionary *Dictionary) Scan(data interface{}) error {
	src := []byte{}
	if data == nil {
		return nil
	}

	switch raw := data.(type) {
	case []byte:
		src = raw
	default:
		return fmt.Errorf("Dictionary: cannot scan from %T", data)
	}

	if len(src) == 0 {
		*dictionary = NewDictionary()
		return nil
	}

	if err := json.Unmarshal(src, dictionary); err != nil {
		return err
	}

	return nil
}

func (dictionary Dictionary) Value() (driver.Value, error) {
	data, err := json.Marshal(dictionary)
	return data, err
}

func (dictionary Dictionary) MarshalJSON() ([]byte, error) {
	return json.Marshal((map[string]interface{})(dictionary))
}

func (dictionary *Dictionary) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, (*map[string]interface{})(dictionary))
}
