package domain

import (
	"fmt"
	"time"
)

const ISODateFormat = "2006-01-02 15:04:05 -0700"

type ISODate time.Time

func (self ISODate) MarshalJSON() ([]byte, error) {
	formatted := (time.Time)(self).Format(ISODateFormat)
	return []byte(fmt.Sprintf("%q", formatted)), nil
}

func (self *ISODate) UnmarshalJSON(data []byte) error {
	if data[0] == byte('"') {
		data = data[1:]
	}

	if data[len(data)-1] == byte('"') {
		data = data[0 : len(data)-1]
	}

	value, err := time.Parse(ISODateFormat, string(data))
	if err != nil {
		return err
	}

	*self = ISODate(value)
	return nil
}
