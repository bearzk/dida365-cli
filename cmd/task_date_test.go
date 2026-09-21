package cmd

import (
	"fmt"
	"testing"
	"time"
)

func localOffset() string {
	_, offset := time.Now().Zone()
	h, m := offset/3600, (offset%3600)/60
	if m < 0 {
		m = -m
	}
	return fmt.Sprintf("%+03d%02d", h, m)
}

func TestNormalizeDateInput(t *testing.T) {
	tz := localOffset()
	tests := []struct {
		name       string
		input      string
		want       string
		wantAllDay bool
		wantErr    bool
	}{
		{
			name:       "date only becomes all day",
			input:      "2026-04-30",
			want:       "2026-04-30T00:00:00" + tz,
			wantAllDay: true,
		},
		{
			name:       "local datetime with space",
			input:      "2026-04-30 18:30",
			want:       "2026-04-30T18:30:00" + tz,
			wantAllDay: false,
		},
		{
			name:       "rfc3339 with timezone",
			input:      "2026-04-30T23:59:59+08:00",
			want:       "2026-04-30T23:59:59+0800",
			wantAllDay: false,
		},
		{
			name:    "reject unsupported format",
			input:   "this saturday",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotAllDay, err := normalizeDateInput(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("normalizeDateInput() = %q, want %q", got, tt.want)
			}
			if gotAllDay != tt.wantAllDay {
				t.Fatalf("normalizeDateInput() allDay = %v, want %v", gotAllDay, tt.wantAllDay)
			}
		})
	}
}