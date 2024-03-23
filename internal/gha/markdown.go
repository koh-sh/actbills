package gha

import (
	"fmt"
	"os"
)

// getOutputPath returns the value of the environment variable GITHUB_STEP_SUMMARY.
// If the environment variable is not set, it returns "/dev/stdout".
func getOutputPath() string {
	outputPath := os.Getenv("GITHUB_STEP_SUMMARY")
	if outputPath == "" {
		outputPath = "/dev/stdout"
	}
	return outputPath
}

// appendToFile appends the given content to the file specified by the filePath.
// If the file does not exist, it will be created.
// If the file exists, the content will be appended to the end of the file.
// The function returns an error if the file cannot be opened or written to.
func appendToFile(filePath, content string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filePath, err)
	}

	return nil
}
