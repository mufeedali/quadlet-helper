package systemd

import (
	"errors"
	"slices"
	"testing"
)

func TestParseIsActiveResult(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		err        error
		wantActive bool
		wantErr    bool
	}{
		{name: "active", output: "active\n", wantActive: true},
		{name: "inactive", output: "inactive\n", err: errors.New("exit status 3")},
		{name: "failed", output: "failed\n", err: errors.New("exit status 3")},
		{name: "bus error", output: "Failed to connect to bus: No medium found\n", err: errors.New("exit status 1"), wantErr: true},
		{name: "unexpected status", output: "weird-status\n", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotActive, err := parseIsActiveResult("demo.service", tt.output, tt.err)
			if gotActive != tt.wantActive {
				t.Fatalf("parseIsActiveResult() active = %v, want %v", gotActive, tt.wantActive)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseIsActiveResult() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseIsActiveMultipleResult(t *testing.T) {
	tests := []struct {
		name    string
		units   []string
		output  string
		err     error
		want    []bool
		wantErr bool
	}{
		{
			name:   "mixed known statuses",
			units:  []string{"one.service", "two.service"},
			output: "active\ninactive\n",
			err:    errors.New("exit status 3"),
			want:   []bool{true, false},
		},
		{
			name:    "bus failure",
			units:   []string{"one.service", "two.service"},
			output:  "Failed to connect to bus: No medium found\n",
			err:     errors.New("exit status 1"),
			wantErr: true,
		},
		{
			name:    "unexpected line count",
			units:   []string{"one.service", "two.service"},
			output:  "active\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIsActiveMultipleResult(tt.units, tt.output, tt.err)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("parseIsActiveMultipleResult() = %v, want %v", got, tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseIsActiveMultipleResult() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
