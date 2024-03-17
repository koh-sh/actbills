package gha

import (
	"fmt"
	"os"
	"sort"
)

const (
	title = "Billable time for workflows in this billable cycle"
	note  = `Please note the following:

- This list shows the execution time for each Workflow at the time this Action was executed.
- Workflows that have been deleted at the time of execution will not be listed.
- Execution times using Larger runners are not included in the aggregation.`
	tableHeader    = "| Workflow | Ubuntu (min) | Windows (min) | Macos (min) |\n"
	tableSeparator = "| --- | --- | --- | --- |\n"
)

// generateMarkdownText generates a markdown-formatted text based on the provided WorkflowBillableTime data.
// It includes a title, a table of billable times for each workflow, and a note.
func generateMarkdownText(wbt WorkflowBillableTime) string {
	markdown := fmt.Sprintf("# %s\n\n", title)
	markdown += generateMarkdownTable(wbt)
	markdown += fmt.Sprintf("\n%s\n", note)

	return markdown
}

// generateMarkdownTable generates a markdown-formatted table of billable times for each workflow.
// The table includes the workflow name and the billable times for Ubuntu, Windows, and macOS.
func generateMarkdownTable(wbt WorkflowBillableTime) string {
	var table string
	table += tableHeader
	table += tableSeparator

	var workflowNames []string
	for name := range wbt {
		workflowNames = append(workflowNames, name)
	}
	sort.Strings(workflowNames)

	for _, name := range workflowNames {
		val := wbt[name]
		table += fmt.Sprintf("| %s | %d | %d | %d |\n", name, val.Ubuntu, val.Windows, val.Macos)
	}

	return table
}

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
