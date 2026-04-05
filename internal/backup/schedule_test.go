package backup

import "testing"

func TestParseSchedule(t *testing.T) {
	tests := []struct {
		name     string
		schedule string
		want     string
	}{
		{name: "daily with time", schedule: "daily 03:15", want: "*-*-* 03:15:00"},
		{name: "weekly shortcut", schedule: "weekly sun 03:00", want: "Sun *-*-* 03:00:00"},
		{name: "passthrough", schedule: "Mon *-*-* 04:30:00", want: "mon *-*-* 04:30:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSchedule(tt.schedule)
			if err != nil {
				t.Fatalf("ParseSchedule() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseSchedule() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseScheduleRejectsInvalidInput(t *testing.T) {
	if _, err := ParseSchedule("not-a-schedule"); err == nil {
		t.Fatal("ParseSchedule() error = nil, want non-nil")
	}
}
