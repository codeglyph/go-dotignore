package dotignore_test

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/codeglyph/go-dotignore"
)

func ExampleNewPatternMatcher() {
	patterns := []string{"*.log", "!important.log", "temp/"}
	matcher, err := dotignore.NewPatternMatcher(patterns)
	if err != nil {
		log.Fatalf("Failed to create pattern matcher: %v", err)
	}

	file := "debug.log"
	matches, _ := matcher.Matches(file)

	fmt.Printf("%s matches: %v\n", file, matches)

	importantFile := "important.log"
	matches, _ = matcher.Matches(importantFile)

	fmt.Printf("%s matches: %v\n", importantFile, matches)
	// Output:
	// debug.log matches: true
	// important.log matches: false
}

func ExamplePatternMatcher_Matches() {
	patterns := []string{"*.txt", "reports/"}
	matcher, err := dotignore.NewPatternMatcher(patterns)
	if err != nil {
		log.Fatalf("Failed to create pattern matcher: %v", err)
	}

	files := []string{"notes.txt", "data.json", "reports/summary.pdf", "images/picture.jpg"}

	for _, file := range files {
		matches, err := matcher.Matches(file)
		if err != nil {
			log.Printf("Error matching file %s: %v", file, err)
			continue
		}

		fmt.Printf("%s matches: %v\n", file, matches)
	}
	// Output:
	// notes.txt matches: true
	// data.json matches: false
	// reports/summary.pdf matches: true
	// images/picture.jpg matches: false
}

func ExampleNewPatternMatcherFromReader() {
	reader := strings.NewReader("*.log\n!important.log\ntemp/")
	matcher, err := dotignore.NewPatternMatcherFromReader(reader)
	if err != nil {
		log.Fatalf("Failed to create pattern matcher: %v", err)
	}

	file := "debug.log"
	matches, _ := matcher.Matches(file)

	fmt.Printf("%s matches: %v\n", file, matches)

	importantFile := "important.log"
	matches, _ = matcher.Matches(importantFile)

	fmt.Printf("%s matches: %v\n", importantFile, matches)
	// Output:
	// debug.log matches: true
	// important.log matches: false
}

func ExampleNewPatternMatcherFromFile() {
	// Create a temporary file to simulate the test.gitignore file
	fileContent := "*.log\n!important.log\ntemp/"
	fileName := "test.gitignore"
	err := os.WriteFile(fileName, []byte(fileContent), 0644)
	if err != nil {
		log.Fatalf("Failed to create test.gitignore file: %v", err)
	}
	defer os.Remove(fileName) // Ensure the file is cleaned up after the test

	matcher, err := dotignore.NewPatternMatcherFromFile(fileName)
	if err != nil {
		log.Fatalf("Failed to create pattern matcher from file: %v", err)
	}

	file := "debug.log"
	matches, _ := matcher.Matches(file)

	fmt.Printf("%s matches: %v\n", file, matches)

	importantFile := "important.log"
	matches, _ = matcher.Matches(importantFile)

	fmt.Printf("%s matches: %v\n", importantFile, matches)
	// Output:
	// debug.log matches: true
	// important.log matches: false
}
