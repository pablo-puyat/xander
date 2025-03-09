package parse

import (
	"testing"
)

func TestParseComicFilename(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		wantSeries     string
		wantIssue      string
		wantYear       string
		wantPublisher  string
		shouldErr      bool
	}{
		{
			name:          "Basic format with cbz extension",
			filename:      "Batman (2016) #001.cbz",
			wantSeries:    "Batman",
			wantIssue:     "001",
			wantYear:      "2016",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Basic format with no extension",
			filename:      "Batman (2016) #001",
			wantSeries:    "Batman",
			wantIssue:     "001",
			wantYear:      "2016",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Basic format with txt extension",
			filename:      "Batman (2016) #001.txt",
			wantSeries:    "Batman",
			wantIssue:     "001",
			wantYear:      "2016",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Format with publisher",
			filename:      "DC Comics - Batman (2016) #001.cbr",
			wantSeries:    "Batman",
			wantIssue:     "001",
			wantYear:      "2016",
			wantPublisher: "DC Comics",
			shouldErr:     false,
		},
		{
			name:          "Format with spaces in series name",
			filename:      "Batman Beyond (2016) #001.cbz",
			wantSeries:    "Batman Beyond",
			wantIssue:     "001",
			wantYear:      "2016",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Format with volume",
			filename:      "Amazing Spider-Man Vol. 5 (2018) #001.cbz",
			wantSeries:    "Amazing Spider-Man",
			wantIssue:     "001",
			wantYear:      "2018",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Invalid format",
			filename:      "random-text",
			shouldErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSeries, gotIssue, gotYear, gotPublisher, err := ParseComicFilename(tt.filename)
			
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
			
			if gotPublisher != tt.wantPublisher {
				t.Errorf("ParseComicFilename() publisher = %v, want %v", gotPublisher, tt.wantPublisher)
			}
		})
	}
}

func TestIsComicFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "CBZ file",
			filename: "Batman #001.cbz",
			want:     true,
		},
		{
			name:     "CBR file",
			filename: "Batman #001.cbr",
			want:     true,
		},
		{
			name:     "Upper case extension",
			filename: "Batman #001.CBZ",
			want:     true,
		},
		{
			name:     "Text file",
			filename: "Batman #001.txt",
			want:     false,
		},
		{
			name:     "No extension",
			filename: "Batman #001",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsComicFile(tt.filename); got != tt.want {
				t.Errorf("IsComicFile() = %v, want %v", got, tt.want)
			}
		})
	}
}