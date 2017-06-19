package domain

import (
	"fmt"
	"strings"
)

type Currency string

var (
	EUR = Currency("EUR")
	USD = Currency("USD")
)

func (self Currency) String() string { return string(self) }

type Money struct {
	Amount   int
	Currency Currency
}

func (self Money) Equal(other *Money) bool {
	return self.Currency == other.Currency && self.Amount == other.Amount
}

func (self Money) Whole() int {
	return self.Amount / 100
}

func (self Money) Cents() int {
	cents := self.Amount % 100
	if cents < 0 {
		return -cents
	}
	return cents
}

func (self Money) String() string {
	sign := ""
	if self.Amount < 0 {
		sign = "-"
	}
	return fmt.Sprintf("%s%d.%02d %s", sign, self.Whole(), self.Cents(), self.Currency)
}

func (self *Money) Scan(src interface{}) error {
	text := ""
	switch input := src.(type) {
	case []byte:
		text = string(input)
	case string:
		text = input
	case nil:
		return nil
	default:
		return fmt.Errorf("cannot scan %T from %T", self, src)
	}

	amountBeforePoint := 0
	amountAfterPoint := 0
	currency := ""
	sign := 1
	if len(text) > 0 && text[0] == '-' {
		sign = -1
		text = text[1:]
	}

	n, err := fmt.Sscanf(text, "%d.%2d %3s", &amountBeforePoint, &amountAfterPoint, &currency)
	if err != nil {
		return err
	}

	if n < 3 {
		return fmt.Errorf("incomplete amount of money: %q", text)
	}

	if amountAfterPoint < 0 {
		return fmt.Errorf("invalid amount: %q", text)
	}

	self.Amount = sign * (amountBeforePoint*100 + amountAfterPoint)
	self.Currency = Currency(strings.ToUpper(currency))

	return nil
}

func (self Money) Value() interface{} {
	return self.String()
}

func (self Money) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", self)), nil
}

func (self *Money) UnmarshalJSON(src []byte) error {
	return self.Scan(src[1:len(src)])
}
