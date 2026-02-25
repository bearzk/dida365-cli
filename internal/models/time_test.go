package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFlexTimeUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantUTC string // expected time in RFC3339 UTC
	}{
		{
			name:    "standard RFC3339Nano with Z",
			input:   `"2026-02-25T20:28:56.267Z"`,
			wantUTC: "2026-02-25T20:28:56.267Z",
		},
		{
			name:    "RFC3339 with colon offset",
			input:   `"2026-02-25T20:28:56.267+00:00"`,
			wantUTC: "2026-02-25T20:28:56.267Z",
		},
		{
			name:    "ISO 8601 basic offset without colon (Dida365 API format)",
			input:   `"2026-02-25T20:28:56.267+0000"`,
			wantUTC: "2026-02-25T20:28:56.267Z",
		},
		{
			name:    "ISO 8601 basic negative offset",
			input:   `"2026-02-25T15:28:56.267-0500"`,
			wantUTC: "2026-02-25T20:28:56.267Z",
		},
		{
			name:    "null JSON value",
			input:   `null`,
			wantErr: false,
			wantUTC: "",
		},
		{
			name:    "invalid format",
			input:   `"not-a-date"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ft FlexTime
			err := json.Unmarshal([]byte(tt.input), &ft)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantUTC == "" {
				// null input: zero value expected
				if !ft.IsZero() {
					t.Errorf("expected zero time, got %v", ft.Time)
				}
				return
			}

			got := ft.UTC().Format(time.RFC3339Nano)
			if got != tt.wantUTC {
				t.Errorf("got %s, want %s", got, tt.wantUTC)
			}
		})
	}
}
