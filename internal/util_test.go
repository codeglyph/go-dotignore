package internal

import (
	"bytes"
	"strings"
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
		{
			name:       "Lines without final newline",
			input:      "line1\nline2",
			expected:   []string{"line1", "line2"},
			shouldFail: false,
		},
		{
			name:       "Single line without newline",
			input:      "single line",
			expected:   []string{"single line"},
			shouldFail: false,
		},
		{
			name:       "Mixed line endings",
			input:      "line1\nline2\rline3\r\nline4",
			expected:   []string{"line1", "line2\rline3", "line4"},
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
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(lines) != len(test.expected) {
				t.Errorf("Expected %d lines, got %d", len(test.expected), len(lines))
				return
			}

			for i, expected := range test.expected {
				if i >= len(lines) || lines[i] != expected {
					t.Errorf("Expected line %d to be %q, got %q", i, expected, lines[i])
				}
			}
		})
	}
}

func TestReadLinesNilReader(t *testing.T) {
	_, err := ReadLines(nil)
	if err == nil {
		t.Error("Expected error for nil reader")
	}

	expectedMsg := "reader cannot be nil"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain %q, got: %v", expectedMsg, err)
	}
}

func TestBuildRegex(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		shouldPass []string
		shouldFail []string
		wantError  bool
	}{
		{
			name:    "Simple wildcard match",
			pattern: "*.txt",
			shouldPass: []string{
				"file.txt", "a.txt", "log.txt", "test.txt", ".txt",
			},
			shouldFail: []string{
				"file.log", "a/b.txt", "filetxt", "file.txt.bak",
			},
		},
		{
			name:    "Single character match",
			pattern: "file?.txt",
			shouldPass: []string{
				"file1.txt", "fileX.txt", "file_.txt", "filea.txt",
			},
			shouldFail: []string{
				"file.txt", "file12.txt", "file/.txt", "fileAB.txt",
			},
		},
		{
			name:    "Double wildcard match",
			pattern: "**/test",
			shouldPass: []string{
				"test", "dir/test", "a/b/c/test", "deep/nested/path/test",
			},
			shouldFail: []string{
				"testing", "test/file", "dir/testing", "test.txt",
			},
		},
		{
			name:    "Double wildcard with slash",
			pattern: "**/dir/",
			shouldPass: []string{
				"dir/", "path/dir/", "a/b/c/dir/",
			},
			shouldFail: []string{
				"dir", "directory/", "path/directory/", "dir/file",
			},
		},
		{
			name:    "Escaped asterisk",
			pattern: "a\\*b",
			shouldPass: []string{
				"a*b",
			},
			shouldFail: []string{
				"ab", "aXb", "a**b", "axb",
			},
		},
		{
			name:    "Character class",
			pattern: "file[0-9].txt",
			shouldPass: []string{
				"file0.txt", "file5.txt", "file9.txt",
			},
			shouldFail: []string{
				"file.txt", "filea.txt", "file10.txt",
			},
		},
		{
			name:    "Directory pattern",
			pattern: "build/",
			shouldPass: []string{
				"build/",
			},
			shouldFail: []string{
				"build", "building/", "rebuild/", "build.txt",
			},
		},
		{
			name:    "Complex pattern with multiple wildcards",
			pattern: "src/**/test/*.js",
			shouldPass: []string{
				"src/test/app.js", "src/components/test/util.js", "src/a/b/c/test/file.js",
			},
			shouldFail: []string{
				"src/test.js", "src/test/", "test/app.js", "src/test/app.ts",
			},
		},
		{
			name:      "Empty pattern",
			pattern:   "",
			wantError: true,
		},
		{
			name:    "Regex metacharacters",
			pattern: "file.log",
			shouldPass: []string{
				"file.log",
			},
			shouldFail: []string{
				"fileXlog", "file_log", "file.log.bak",
			},
		},
		{
			name:    "Pattern with special characters",
			pattern: "file$(test).txt",
			shouldPass: []string{
				"file$(test).txt",
			},
			shouldFail: []string{
				"fileXtest.txt", "file$test.txt", "file(test).txt",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			regex, err := BuildRegex(test.pattern)

			if test.wantError {
				if err == nil {
					t.Errorf("Expected error for pattern %q, got nil", test.pattern)
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to build regex for pattern %q: %v", test.pattern, err)
			}

			// Test positive matches
			for _, input := range test.shouldPass {
				if !regex.MatchString(input) {
					t.Errorf("Pattern %q should match input %q, but it did not", test.pattern, input)
				}
			}

			// Test negative matches
			for _, input := range test.shouldFail {
				if regex.MatchString(input) {
					t.Errorf("Pattern %q should not match input %q, but it did", test.pattern, input)
				}
			}
		})
	}
}

func TestBuildRegexEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		want    bool
	}{
		{
			name:    "Trailing backslash",
			pattern: "file\\",
			input:   "file\\",
			want:    true,
		},
		{
			name:    "Escaped question mark",
			pattern: "file\\?",
			input:   "file?",
			want:    true,
		},
		{
			name:    "Unmatched bracket",
			pattern: "file[incomplete",
			input:   "file[incomplete",
			want:    true,
		},
		{
			name:    "Double asterisk at end",
			pattern: "dir/**",
			input:   "dir/file.txt",
			want:    true,
		},
		{
			name:    "Single asterisk vs double asterisk",
			pattern: "dir/*",
			input:   "dir/sub/file.txt",
			want:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			regex, err := BuildRegex(test.pattern)
			if err != nil {
				t.Fatalf("Failed to build regex: %v", err)
			}

			got := regex.MatchString(test.input)
			if got != test.want {
				t.Errorf("Pattern %q matching %q: expected %v, got %v", test.pattern, test.input, test.want, got)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Forward slashes",
			input:    "path/to/file",
			expected: "path/to/file",
		},
		{
			name:     "Backward slashes",
			input:    "path\\to\\file",
			expected: "path/to/file",
		},
		{
			name:     "Mixed slashes",
			input:    "path/to\\file",
			expected: "path/to/file",
		},
		{
			name:     "Double slashes",
			input:    "path//to///file",
			expected: "path/to/file",
		},
		{
			name:     "Leading dot slash",
			input:    "./path/to/file",
			expected: "path/to/file",
		},
		{
			name:     "Trailing slash",
			input:    "path/to/file/",
			expected: "path/to/file",
		},
		{
			name:     "Root path",
			input:    "/",
			expected: "/",
		},
		{
			name:     "Empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "Current directory",
			input:    "./",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := NormalizePath(test.input)
			if result != test.expected {
				t.Errorf("NormalizePath(%q) = %q, want %q", test.input, result, test.expected)
			}
		})
	}
}

func BenchmarkBuildRegex(b *testing.B) {
	patterns := []string{
		"*.txt",
		"**/*.js",
		"src/**/test/*.go",
		"build/",
		"node_modules/**",
		"*.{log,tmp,cache}",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pattern := range patterns {
			_, err := BuildRegex(pattern)
			if err != nil {
				b.Fatalf("BuildRegex failed: %v", err)
			}
		}
	}
}

func BenchmarkRegexMatch(b *testing.B) {
	regex, err := BuildRegex("**/*.js")
	if err != nil {
		b.Fatalf("BuildRegex failed: %v", err)
	}

	testPaths := []string{
		"app.js",
		"src/app.js",
		"src/components/Header.js",
		"build/static/js/main.js",
		"node_modules/react/index.js",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			regex.MatchString(path)
		}
	}
}
