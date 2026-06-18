package cmd

import "testing"

func TestnormalizeDateInput(t *testing.T) {
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
			want:       "2026-04-30T00:00:00+0800",
			wantAllDay: true,
		},
		{
			name:       "local datetime with space",
			input:      "2026-04-30 18:30",
			want:       "2026-04-30T18:30:00+0800",
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