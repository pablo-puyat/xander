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
			name:          "Modern format with digital and group tags",
			filename:      "Zawa + The Belly of the Beast (2024) (digital) (Son of Ultron-Empire).cbr",
			wantSeries:    "Zawa + The Belly of the Beast",
			wantIssue:     "1",
			wantYear:      "2024",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Year as title format",
			filename:      "1949 (2024) (Digital) (DR & Quinch-Empire).cbr",
			wantSeries:    "1949",
			wantIssue:     "1",
			wantYear:      "2024",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Year-month pattern",
			filename:      "2007-04 - But the Past Ain't through with You (Digital-1280) (Empire).cbr",
			wantSeries:    "But the Past Ain't through with You",
			wantIssue:     "04",
			wantYear:      "2007",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Series with issue number",
			filename:      "Absolute Batman 001 (2024) (Webrip) (The Last Kryptonian-DCP).cbr",
			wantSeries:    "Absolute Batman",
			wantIssue:     "001",
			wantYear:      "2024",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Series with issue number 2",
			filename:      "Absolute Superman 003 (2025) (Digital) (Pyrate-DCP).cbz",
			wantSeries:    "Absolute Superman",
			wantIssue:     "003",
			wantYear:      "2025",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Title with dash and issue number",
			filename:      "Batman - The Long Halloween - The Last Halloween 000 (2024) (Webrip) (The Last Kryptonian-DCP).cbr",
			wantSeries:    "Batman",
			wantIssue:     "000",
			wantYear:      "2024",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Volume pattern",
			filename:      "Black Panther v01 - A Nation Under Our Feet (2016) (Digital) (F) (Asgard-Empire).cbr",
			wantSeries:    "Black Panther",
			wantIssue:     "01",
			wantYear:      "2016",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Simple title with apostrophe",
			filename:      "DC's Lex and the City 001 (2025) (Webrip) (The Last Kryptonian-DCP).cbr",
			wantSeries:    "DC's Lex and the City",
			wantIssue:     "001",
			wantYear:      "2025",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Volume with year",
			filename:      "G.I. Joe - Cobra v01 (2009) (Minutemen-DarthTremens) (RC).cbz",
			wantSeries:    "G.I. Joe - Cobra",
			wantIssue:     "01",
			wantYear:      "2009",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Issue (of total) format",
			filename:      "Jim Henson's Labyrinth 01 (of 08) (2024) (digital) (Son of Ultron-Empire).cbr",
			wantSeries:    "Jim Henson's Labyrinth",
			wantIssue:     "01",
			wantYear:      "2024",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Range of issues",
			filename:      "Paklis 001-005 (2017) (Digital) (Zone-Empire)",
			wantSeries:    "Paklis 001-005",
			wantIssue:     "1",
			wantYear:      "2017",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Ultimate series",
			filename:      "Ultimate X-Men 007 (2024) (Digital) (Shan-Empire).cbz",
			wantSeries:    "Ultimate X-Men",
			wantIssue:     "007",
			wantYear:      "2024",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Simple series title",
			filename:      "Angel and the Ape 001.cbz",
			wantSeries:    "Angel and the Ape",
			wantIssue:     "001",
			wantYear:      "",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Dot-separated format",
			filename:      "G.I.JOE.Cobra.Snake.Eyes.and.Storm.Shadow.December.2012.RETAiL.COMiC.eBOOk-rebOOk.pdf.pdf",
			wantSeries:    "G I JOE Cobra Snake Eyes and Storm Shadow",
			wantIssue:     "1",
			wantYear:      "2012",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Series with duplicate number",
			filename:      "Absolute Wonder Woman 003 (2025) (Digital) (Pyrate-DCP) 2.cbz",
			wantSeries:    "Absolute Wonder Woman",
			wantIssue:     "003",
			wantYear:      "2025",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Hyphenated title with year",
			filename:      "Ultimate Universe - One Year In 001 (2025) (Digital) (Shan-Empire).cbz",
			wantSeries:    "Ultimate Universe",
			wantIssue:     "001",
			wantYear:      "2025",
			wantPublisher: "",
			shouldErr:     false,
		},
		{
			name:          "Invalid format",
			filename:      "",
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