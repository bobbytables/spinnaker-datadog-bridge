package types

import (
	"fmt"
	"strconv"
	"time"
)

// Timestamp implements JSON marshalling / unmarshalling to allow
// using it for unix timestamp values in struct values
type Timestamp struct {
	time.Time
}

// MarshalJSON implements json.Marshaler
func (t *Timestamp) MarshalJSON() ([]byte, error) {
	ts := t.Time.Unix()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	ts, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}

	t.Time = time.Unix(int64(ts)/1000, 0)

	return nil
}
