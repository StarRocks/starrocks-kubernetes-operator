package main

import (
	"fmt"
	"os"
	"regexp"
)

func main() {
	// Input Markdown file
	inputFile := "../CHANGELOG.md"

	// GitHub project URL
	githubURL := "https://github.com/StarRocks/starrocks-kubernetes-operator/pull"

	// Read the input file
	content, err := os.ReadFile(inputFile)
	if err != nil {
		panic(err)
	}

	// Create a regular expression to match (#number)
	re := regexp.MustCompile(`\(#([0-9]+)\)`)

	// Replace (#number) with [#number](githubURL/number)
	newContent := re.ReplaceAllStringFunc(string(content), func(s string) string {
		number := s[2 : len(s)-1] // Extract the number
		return fmt.Sprintf("[#%s](%s/%s)", number, githubURL, number)
	})

	// Write the new content back to the input file
	var perm os.FileMode = 0600
	err = os.WriteFile(inputFile, []byte(newContent), perm)
	if err != nil {
		panic(err)
	}
}
