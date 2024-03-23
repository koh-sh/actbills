package bills

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-github/v60/github"
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

// WorkflowBillableTimes represents a map of workflow names to their corresponding WorkflowBillableTime
type WorkflowBillableTimes map[string]WorkflowBillableTime

// WorkflowBillableTime represents the total billable time for each environment in a workflow
type WorkflowBillableTime struct {
	Ubuntu  int64 // Total billable time for the Ubuntu environment (in minutes)
	Windows int64 // Total billable time for the Windows environment (in minutes)
	Macos   int64 // Total billable time for the Mac environment (in minutes)
}

// GenerateMarkdownReport generates a markdown-formatted report based on the provided WorkflowBillableTimes data.
// It includes a title, a table of billable times for each workflow, and a note.
func (w WorkflowBillableTimes) generateMarkdownReport() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))
	sb.WriteString(w.generateMarkdownTable())
	sb.WriteString(w.calculateTotal().formatBoldMarkdownRow("Total"))
	sb.WriteString(fmt.Sprintf("\n%s\n", note))

	return sb.String()
}

// calculateTotal calculates the total billable time for each environment across all workflows
func (w WorkflowBillableTimes) calculateTotal() WorkflowBillableTime {
	var totalBillableTime WorkflowBillableTime
	for _, billableTime := range w {
		totalBillableTime.Ubuntu += billableTime.Ubuntu
		totalBillableTime.Windows += billableTime.Windows
		totalBillableTime.Macos += billableTime.Macos
	}
	return totalBillableTime
}

// generateMarkdownTable generates a markdown-formatted table of billable times for each workflow.
// The table includes the workflow name and the billable times for Ubuntu, Windows, and macOS.
func (w WorkflowBillableTimes) generateMarkdownTable() string {
	var sb strings.Builder
	sb.WriteString(tableHeader)
	sb.WriteString(tableSeparator)

	workflowNames := w.sortWorkflowNames()

	for _, name := range workflowNames {
		sb.WriteString(w[name].formatMarkdownRow(name))
	}

	return sb.String()
}

// sortWorkflowNames returns a sorted slice of workflow names.
func (w WorkflowBillableTimes) sortWorkflowNames() []string {
	var workflowNames []string
	for name := range w {
		workflowNames = append(workflowNames, name)
	}
	sort.Strings(workflowNames)
	return workflowNames
}

// formatMarkdownRow formats the billable time for each environment as a markdown table row
func (e WorkflowBillableTime) formatMarkdownRow(title string) string {
	return fmt.Sprintf("| %s | %d | %d | %d |\n", title, e.Ubuntu, e.Windows, e.Macos)
}

// formatBoldMarkdownRow formats the billable time for each environment as a bold markdown table row
func (e WorkflowBillableTime) formatBoldMarkdownRow(title string) string {
	return fmt.Sprintf("| **%s** | **%d** | **%d** | **%d** |\n", title, e.Ubuntu, e.Windows, e.Macos)
}

// CreateReport retrieves billable time for workflows and generates a markdown report
func CreateReport(repository string) error {
	owner, repo, err := extractOwnerAndRepo(repository)
	if err != nil {
		return err
	}

	client := createGitHubClient()
	workflows, err := fetchWorkflows(client, owner, repo)
	if err != nil {
		return err
	}

	wbt, err := generateWorkflowBillableTimes(client, owner, repo, workflows)
	if err != nil {
		return err
	}
	err = appendToFile(getOutputPath(), wbt.generateMarkdownReport())
	if err != nil {
		return err
	}

	return nil
}

// generateWorkflowBillableTime generates a WorkflowBillableTimes for the specified workflows
func generateWorkflowBillableTimes(client *github.Client, owner, repo string, workflows []*github.Workflow) (WorkflowBillableTimes, error) {
	wbt := make(WorkflowBillableTimes)

	for _, workflow := range workflows {
		billMap, err := fetchWorkflowBillMap(client, owner, repo, *workflow.ID)
		if err != nil {
			return nil, err
		}

		wbt[*workflow.Name] = WorkflowBillableTime{
			Ubuntu:  getMinutesForEnv(billMap, "UBUNTU"),
			Windows: getMinutesForEnv(billMap, "WINDOWS"),
			Macos:   getMinutesForEnv(billMap, "MACOS"),
		}
	}

	return wbt, nil
}
