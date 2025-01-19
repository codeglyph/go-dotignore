package internal

import (
	"bytes"
	"testing"
)

func TestReadLines(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expected   []string
		shouldFail bool
	}{
		{
			name:       "Simple lines",
			input:      "line1\nline2\nline3\n",
			expected:   []string{"line1", "line2", "line3"},
			shouldFail: false,
		},
		{
			name:       "Lines with UTF-8 BOM",
			input:      string([]byte{0xEF, 0xBB, 0xBF}) + "line1\nline2\n",
			expected:   []string{"line1", "line2"},
			shouldFail: false,
		},
		{
			name:       "Empty input",
			input:      "",
			expected:   []string{},
			shouldFail: false,
		},
		{
			name:       "Input with only whitespace lines",
			input:      "\n  \n\n",
			expected:   []string{"", "  ", ""},
			shouldFail: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte(test.input))
			lines, err := ReadLines(reader)
			if test.shouldFail {
				if err == nil {
					t.Errorf("Expected an error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(lines) != len(test.expected) {
					t.Errorf("Expected %d lines, got %d", len(test.expected), len(lines))
				}
				for i, line := range test.expected {
					if lines[i] != line {
						t.Errorf("Expected line %d to be %q, got %q", i, line, lines[i])
					}
				}
			}
		})
	}
}

func TestBuildRegex(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		shouldPass []string
		shouldFail []string
	}{
		{
			name:    "Wildcard match",
			pattern: "*.txt",
			shouldPass: []string{
				"file.txt", "a.txt", "log.txt",
			},
			shouldFail: []string{
				"file.log", "a/b.txt", "filetxt",
			},
		},
		{
			name:    "Single character match",
			pattern: "file?.txt",
			shouldPass: []string{
				"file1.txt", "fileX.txt", "file_.txt",
			},
			shouldFail: []string{
				"file.txt", "file12.txt", "file/.txt",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			regex, err := BuildRegex(test.pattern)
			if err != nil {
				t.Fatalf("Failed to build regex: %v", err)
			}

			for _, input := range test.shouldPass {
				if !regex.MatchString(input) {
					t.Errorf("Expected pattern %q to match input %q, but it did not", test.pattern, input)
				}
			}

			for _, input := range test.shouldFail {
				if regex.MatchString(input) {
					t.Errorf("Expected pattern %q not to match input %q, but it did", test.pattern, input)
				}
			}
		})
	}
}
