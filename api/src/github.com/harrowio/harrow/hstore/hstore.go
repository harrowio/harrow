package hstore

import (
	"database/sql"
	"database/sql/driver"
	"strings"
)

func Quote(s interface{}) string {
	var str string
	switch v := s.(type) {
	case sql.NullString:
		if !v.Valid {
			return "NULL"
		}
		str = v.String
	case string:
		str = v
	default:
		panic("not a string or sql.NullString")
	}

	str = strings.Replace(str, "\n", "\\n", -1)
	str = strings.Replace(str, "\\", "\\\\", -1)
	return `"` + strings.Replace(str, "\"", "\\\"", -1) + `"`
}

func Value(value map[string]string) (driver.Value, error) {
	if value == nil {
		return nil, nil
	}
	parts := []string{}
	for key, val := range value {
		thispart := Quote(key) + "=>" + Quote(val)
		parts = append(parts, thispart)
	}
	return []byte(strings.Join(parts, ",")), nil
}

func Scan(value interface{}) (map[string]string, error) {
	if value == nil {
		return nil, nil
	}
	var m map[string]string = make(map[string]string)
	var b byte
	pair := [][]byte{{}, {}}
	pairIndex := 0
	inQuote := false
	didQuote := false
	sawSlash := false
	bindex := 0
	for bindex, b = range value.([]byte) {
		if sawSlash {
			pair[pairIndex] = append(pair[pairIndex], b)
			sawSlash = false
			continue
		}

		switch b {
		case '\\':
			sawSlash = true
			continue
		case '"':
			inQuote = !inQuote
			if !didQuote {
				didQuote = true
			}
			continue
		default:
			if !inQuote {
				switch b {
				case ' ', '\t', '\n', '\r':
					continue
				case '=':
					continue
				case '>':
					pairIndex = 1
					didQuote = false
					continue
				case ',':
					s := string(pair[1])
					if !didQuote && len(s) == 4 && strings.ToLower(s) == "null" {
						m[string(pair[0])] = ""
					} else {
						m[string(pair[0])] = string(pair[1])
					}
					pair[0] = []byte{}
					pair[1] = []byte{}
					pairIndex = 0
					continue
				}
			}
		}
		pair[pairIndex] = append(pair[pairIndex], b)
	}
	if bindex > 0 {
		s := string(pair[1])
		if !didQuote && len(s) == 4 && strings.ToLower(s) == "null" {
			m[string(pair[0])] = ""
		} else {
			m[string(pair[0])] = string(pair[1])
		}
	}
	return m, nil
}
