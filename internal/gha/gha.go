package gha

import (
	"context"
	"fmt"
	"os"
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

// createGitHubClient creates a new GitHub API client with optional authentication token
func createGitHubClient() *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return github.NewClient(nil)
	}
	return github.NewClient(nil).WithAuthToken(token)
}

// fetchWorkflows retrieves a list of workflows for the specified repository
func fetchWorkflows(client *github.Client, owner, repo string) ([]*github.Workflow, error) {
	var allWorkflows []*github.Workflow
	opts := &github.ListOptions{PerPage: 100}

	for {
		workflows, resp, err := client.Actions.ListWorkflows(context.Background(), owner, repo, opts)
		if err != nil {
			return nil, err
		}

		allWorkflows = append(allWorkflows, workflows.Workflows...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allWorkflows, nil
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

// fetchWorkflowBillableTime retrieves the billable time map for a specific workflow
func fetchWorkflowBillMap(client *github.Client, owner, repo string, workflowID int64) (github.WorkflowBillMap, error) {
	usage, _, err := client.Actions.GetWorkflowUsageByID(context.Background(), owner, repo, workflowID)
	if err != nil {
		return nil, err
	}

	return *usage.Billable, nil
}

// extractOwnerAndRepo extracts the owner and repository name from the provided repository argument or environment variable
func extractOwnerAndRepo(repo string) (string, string, error) {
	var ownerRepo string
	if repo != "" {
		ownerRepo = repo
	} else {
		ownerRepo = os.Getenv("GITHUB_REPOSITORY")
	}

	if ownerRepo == "" {
		return "", "", fmt.Errorf("repository name not provided and GITHUB_REPOSITORY environment variable not set")
	}

	parts := strings.Split(ownerRepo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository name format: %s", ownerRepo)
	}

	return parts[0], parts[1], nil
}

// getMinutesForEnv retrieves the total billable time in minutes for a specific environment
func getMinutesForEnv(billMap github.WorkflowBillMap, env string) int64 {
	bill, ok := billMap[env]
	if !ok {
		return 0
	}

	return bill.GetTotalMS() / 60000 // convert milliseconds to minutes
}
