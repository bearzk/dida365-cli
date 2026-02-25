package models

import (
	"fmt"
	"time"
)

// FlexTime is a time.Time that can unmarshal both RFC3339 and ISO 8601 basic
// offset formats (e.g. "+0000" without colon), as returned by the Dida365 API
// for completedTime fields.
type FlexTime struct {
	time.Time
}

// flexTimeLayouts lists formats tried in order during JSON unmarshaling.
var flexTimeLayouts = []string{
	time.RFC3339Nano,                     // "2006-01-02T15:04:05.999999999Z07:00"
	"2006-01-02T15:04:05.999999999-0700", // ISO 8601 basic: +0000 (no colon)
	time.RFC3339,                         // "2006-01-02T15:04:05Z07:00"
	"2006-01-02T15:04:05-0700",           // ISO 8601 basic without sub-seconds
}

// UnmarshalJSON implements json.Unmarshaler. It accepts null (zero value) and
// any of the formats in flexTimeLayouts.
func (ft *FlexTime) UnmarshalJSON(data []byte) error {
	s := string(data)

	if s == "null" {
		return nil
	}

	// Strip surrounding quotes
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return fmt.Errorf("FlexTime: expected JSON string, got %s", s)
	}
	s = s[1 : len(s)-1]

	for _, layout := range flexTimeLayouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			ft.Time = t
			return nil
		}
	}

	return fmt.Errorf("FlexTime: cannot parse %q as a time value", s)
}
