package gha

import (
	"fmt"
	"os"
)

type TableRow struct {
	Environment  string
	TotalMinutes float64
	Cost         float64
}

func GenerateMarkdownText(title string, tableRows []TableRow, note string) string {
	markdown := fmt.Sprintf("# %s\n\n", title)

	// Generate the table
	markdown += "| Environment | Total Minutes | Cost |\n"
	markdown += "| --- | --- | --- |\n"

	for _, row := range tableRows {
		markdown += fmt.Sprintf("| %s | %.2f | $%.2f |\n", row.Environment, row.TotalMinutes, row.Cost)
	}

	// Add the note
	markdown += fmt.Sprintf("\n%s\n", note)

	return markdown
}

// getOutputPath returns the value of the environment variable GITHUB_STEP_SUMMARY
// if it is set, or "/dev/stdout" if the environment variable is not set.
func getOutputPath() string {
	outputPath := os.Getenv("GITHUB_STEP_SUMMARY")
	if outputPath == "" {
		outputPath = "/dev/stdout"
	}
	return outputPath
}

// appendToFile appends the given string to the file specified by the file path.
// If the file does not exist, it will be created.
// If the file exists, the string will be appended to the end of the file.
// The function returns an error if the file cannot be opened or written to.
func appendToFile(filePath, content string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
