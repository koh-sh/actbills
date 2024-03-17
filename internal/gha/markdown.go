package gha

import (
	"fmt"
	"os"
	"sort"
)

func generateMarkdownText(wbt WorkflowBillableTime) string {
	title := "Billable time for each Workflows in this billable cycle"
	note := "hoge"

	markdown := fmt.Sprintf("# %s\n\n", title)

	// Generate the table
	markdown += "| Workflow | Ubuntu (min) | Windows (min) | Macos (min) |\n"
	markdown += "| --- | --- | --- | --- |\n"

	// Sort the workflow names
	var workflowNames []string
	for name := range wbt {
		workflowNames = append(workflowNames, name)
	}
	sort.Strings(workflowNames)
	// Print the data in the sorted order
	for _, name := range workflowNames {
		val := wbt[name]
		markdown += fmt.Sprintf("| %s | %d | %d | %d |\n", name, val.Ubuntu, val.Windows, val.Macos)
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
