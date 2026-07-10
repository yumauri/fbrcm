package jsonscan

import "testing"

func TestScanStringEnd(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		start   int
		wantEnd int
		wantOK  bool
	}{
		{
			name:    "empty string",
			body:    `""`,
			start:   0,
			wantEnd: 1,
			wantOK:  true,
		},
		{
			name:    "simple string",
			body:    `"hello"`,
			start:   0,
			wantEnd: 6,
			wantOK:  true,
		},
		{
			name:    "escaped quote",
			body:    `"say \"hi\""`,
			start:   0,
			wantEnd: 11,
			wantOK:  true,
		},
		{
			name:    "escaped backslash",
			body:    `"path\\to"`,
			start:   0,
			wantEnd: 9,
			wantOK:  true,
		},
		{
			name:    "mid object key",
			body:    `{"key":"value"}`,
			start:   1,
			wantEnd: 5,
			wantOK:  true,
		},
		{
			name:   "not a quote",
			body:   `123`,
			start:  0,
			wantOK: false,
		},
		{
			name:   "unterminated",
			body:   `"open`,
			start:  0,
			wantOK: false,
		},
		{
			name:   "start out of range",
			body:   `"x"`,
			start:  5,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			end, ok := ScanStringEnd([]byte(tt.body), tt.start)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && end != tt.wantEnd {
				t.Fatalf("end = %d, want %d", end, tt.wantEnd)
			}
		})
	}
}
