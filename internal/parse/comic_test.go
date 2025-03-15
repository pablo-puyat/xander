package parse

import (
	"testing"
)

func TestParseComicFilename(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		wantSeries    string
		wantIssue     string
		wantYear      string
		wantPublisher string
		shouldErr     bool
	}{
		{
			name:       "Common format",
			filename:   "Absolute Superman 003 (2025) (Digital) (Pyrate-DCP).cbz",
			wantSeries: "Absolute Superman",
			wantIssue:  "003",
			wantYear:   "2025",
			shouldErr:  false,
		},
		{
			name:       "Common format",
			filename:   "You Look Like Death -Tales from the Umbrella Academy (2021) (digital) (Son of Ultron-Empire).cbr",
			wantSeries: "You Look Like Death -Tales from the Umbrella Academy",
			wantIssue:  "",
			wantYear:   "2021",
			shouldErr:  false,
		},
		{
			name:       "Common format",
			filename:   "Wolverine - Deep Cut 001 (2024) (digital) (Marika-Empire).cbz",
			wantSeries: "Wolverine - Deep Cut",
			wantIssue:  "001",
			wantYear:   "2024",
			shouldErr:  false,
		},
		{
			name:      "Invalid format",
			filename:  "",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSeries, gotIssue, gotYear, err := ParseComicFilename(tt.filename)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("ParseComicFilename() error = nil, expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseComicFilename() error = %v, expected no error", err)
				return
			}

			if gotSeries != tt.wantSeries {
				t.Errorf("ParseComicFilename() series = %v, want %v", gotSeries, tt.wantSeries)
			}

			if gotIssue != tt.wantIssue {
				t.Errorf("ParseComicFilename() issue = %v, want %v", gotIssue, tt.wantIssue)
			}

			if gotYear != tt.wantYear {
				t.Errorf("ParseComicFilename() year = %v, want %v", gotYear, tt.wantYear)
			}

		})
	}
}
