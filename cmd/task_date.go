package cmd

import (
	"fmt"
	"strings"
	"time"
)

const apiDateTimeLayout = "2006-01-02T15:04:05-0700"

var dueDateLayoutsWithZone = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05.999999999-0700",
	"2006-01-02T15:04:05-0700",
}

var dueDateLayoutsLocal = []string{
	"2006-01-02 15:04",
	"2006-01-02T15:04",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
}

func normalizeDateInput(input string) (string, bool, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", false, fmt.Errorf("due date cannot be empty")
	}

	if t, err := time.ParseInLocation("2006-01-02", trimmed, time.Local); err == nil {
		return t.Format(apiDateTimeLayout), true, nil
	}

	for _, layout := range dueDateLayoutsLocal {
		if t, err := time.ParseInLocation(layout, trimmed, time.Local); err == nil {
			return t.Format(apiDateTimeLayout), false, nil
		}
	}

	for _, layout := range dueDateLayoutsWithZone {
		if t, err := time.Parse(layout, trimmed); err == nil {
			return t.Format(apiDateTimeLayout), false, nil
		}
	}

	return "", false, fmt.Errorf("unsupported due date format %q; use YYYY-MM-DD, YYYY-MM-DD HH:MM, YYYY-MM-DDTHH:MM, or RFC3339", input)
}