package types

import (
	"encoding/json"
	"time"
)
type UnixTime struct {
  time.Time
}

// Override UnmarshalJSON to handle timestamp in unix format
func (t *UnixTime) UnmarshalJSON(b []byte) error {
	var timestamp int64
	err := json.Unmarshal(b, &timestamp)
	if err != nil {
		return err
	}
	t.Time = time.Unix(timestamp, 0)
	return nil
}
