package mailDispatcher

import (
	"testing"
)

func TestPgTime_UnmarshalJSON(t *testing.T) {
	timestamps := []string{
		// timestamp as produced by row_to_json in Postgresql 9.3
		"2014-11-29 18:06:17.640985+00",
		// timestamp as produced by row_to_json in Postgresql 9.4
		"2014-11-29T18:06:17.640985+00:00",
	}

	results := make([]PgTime, len(timestamps))

	for i, ts := range timestamps {
		if err := results[i].UnmarshalJSON([]byte(ts)); err != nil {
			t.Fatal(err)
		}
	}
}
