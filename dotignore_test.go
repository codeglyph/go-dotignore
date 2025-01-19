package dotignore

import (
	"os"
	"testing"
)

func TestNewPatternMatcherFromFile(t *testing.T) {
	// Create a temporary .ignore file
	ignoreContent := `
ignore.exe
6.out
ignore.so
!ok.go
!ok.py
!ok.exe
!ok.a
`
	tempFile, err := os.CreateTemp("", "test.ignore")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	if _, err := tempFile.WriteString(ignoreContent); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tempFile.Close()

	matcher, err := NewPatternMatcherFromFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to parse ignore file: %v", err)
	}

	// Define ignored and included files
	ignored := []string{"ignore.exe", "6.out", "ignore.so"}
	for _, name := range ignored {
		isSkip, err := matcher.Matches(name)
		if err != nil || !isSkip {
			t.Errorf("Expected to ignore %s, got: %v, err: %v", name, isSkip, err)
		}
	}

	included := []string{"ok.go", "ok.py", "ok.exe", "ok.a"}
	for _, name := range included {
		isSkip, err := matcher.Matches(name)
		if err != nil || isSkip {
			t.Errorf("Expected to include %s, got: %v, err: %v", name, isSkip, err)
		}
	}
}

func TestMatches(t *testing.T) {
	patterns := []string{
		"**",
		"**/",
		"dir/**",
		"**/dir2/*",
		"**/dir2/**",
		"**/file",
		"**/**/*.txt",
		"a**/*.txt",
		"a/*.txt",
		"a[b-d]e",
		"abc.def",
		"abc?def",
		"a\\*b",
		"**/foo/bar",
		"abc/**",
	}
	matcher, err := NewPatternMatcher(patterns)
	if err != nil {
		t.Fatalf("Failed to create matcher: %v", err)
	}

	tests := []struct {
		file     string
		expected bool
	}{
		{"file", true},
		{"file/", true},
		{"dir/file", true},
		{"dir/file/", true},
		{"dir/dir2/file", true},
		{"dir/dir2/file/", true},
		{"a/file.txt", true},
		{"a/dir/file.txt", true},
		{"abc.def", true},
		{"a*b", true},
		{"a\\b", true},
		{"foo/bar", true},
		{"abc/def", true},
	}

	for _, test := range tests {
		result, err := matcher.Matches(test.file)
		if err != nil {
			t.Errorf("Error matching file %s: %v", test.file, err)
		}
		if result != test.expected {
			t.Errorf("File %s: expected %v, got %v", test.file, test.expected, result)
		}
	}
}

func TestBuildIgnorePatterns(t *testing.T) {
	patterns := []string{"docs", "config"}
	ignorePatterns, _ := buildIgnorePatterns(patterns)
	if len(ignorePatterns) != 2 {
		t.Errorf("expected 2 element slice, got %v", len(ignorePatterns))
	}
}

func TestBuildIgnorePatternsStripEmptyPatterns(t *testing.T) {
	patterns := []string{"docs", "config", ""}
	ignorePatterns, _ := buildIgnorePatterns(patterns)
	if len(ignorePatterns) != 2 {
		t.Errorf("expected 2 element slice, got %v", len(ignorePatterns))
	}
}

func TestBuildIgnorePatternsExceptionFlag(t *testing.T) {
	patterns := []string{"docs", "!docs/README.md"}
	ignorePatterns, _ := buildIgnorePatterns(patterns)
	if !ignorePatterns[1].Negate {
		t.Errorf("expected negate to be true, got %v", ignorePatterns[1].Negate)
	}
}

func TestBuildIgnorePatternsLeadingSpaceTrimmed(t *testing.T) {
	patterns := []string{"docs", "  !docs/README.md"}
	ignorePatterns, _ := buildIgnorePatterns(patterns)
	if !ignorePatterns[1].Negate {
		t.Errorf("expected negate to be true, got %v", ignorePatterns[1].Negate)
	}
}

func TestBuildIgnorePatternsTrailingSpaceTrimmed(t *testing.T) {
	patterns := []string{"docs", "!docs/README.md  "}
	ignorePatterns, _ := buildIgnorePatterns(patterns)
	if !ignorePatterns[1].Negate {
		t.Errorf("expected negate to be true, got %v", ignorePatterns[1].Negate)
	}
}

func TestBuildIgnorePatternsErrorSingleException(t *testing.T) {
	patterns := []string{"!"}
	_, err := buildIgnorePatterns(patterns)
	if err == nil {
		t.Errorf("expected error on single exclamation point, got %v", err)
	}
}

func TestBuildIgnorePatternsFolderSplit(t *testing.T) {
	patterns := []string{"docs/config/CONFIG.md"}
	ignorePatterns, _ := buildIgnorePatterns(patterns)
	if ignorePatterns[0].ParentDirs[0] != "docs" {
		t.Errorf("expected first element in dirs slice to be docs, got %v", ignorePatterns[0].ParentDirs[0])
	}
	if ignorePatterns[0].ParentDirs[1] != "config" {
		t.Errorf("expected second element in dirs slice to be config, got %v", ignorePatterns[0].ParentDirs[1])
	}
}
