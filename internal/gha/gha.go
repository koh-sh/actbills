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

// WorkflowBillableTime represents a map of workflow names to their corresponding EnvBillableTime
type WorkflowBillableTime map[string]EnvBillableTime

// EnvBillableTime represents the total billable time for each environment in a workflow
type EnvBillableTime struct {
	Ubuntu  int64 // Total billable time for the Ubuntu environment (in minutes)
	Windows int64 // Total billable time for the Windows environment (in minutes)
	Macos   int64 // Total billable time for the Mac environment (in minutes)
}

// generateMarkdownText generates a markdown-formatted text based on the provided WorkflowBillableTime data.
// It includes a title, a table of billable times for each workflow, and a note.
func (w WorkflowBillableTime) generateMarkdownText() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))
	sb.WriteString(w.generateMarkdownTable())
	sb.WriteString(w.total().markdownBoldRow("Total"))
	sb.WriteString(fmt.Sprintf("\n%s\n", note))

	return sb.String()
}

// return WorkflowBillableTime total as EnvBillableTime
func (w WorkflowBillableTime) total() EnvBillableTime {
	var totalBillableTime EnvBillableTime
	for _, billableTime := range w {
		totalBillableTime.Ubuntu += billableTime.Ubuntu
		totalBillableTime.Windows += billableTime.Windows
		totalBillableTime.Macos += billableTime.Macos
	}
	return totalBillableTime
}

// generateMarkdownTable generates a markdown-formatted table of billable times for each workflow.
// The table includes the workflow name and the billable times for Ubuntu, Windows, and macOS.
func (wbt WorkflowBillableTime) generateMarkdownTable() string {
	var sb strings.Builder
	sb.WriteString(tableHeader)
	sb.WriteString(tableSeparator)

	workflowNames := wbt.sortWorkflowNames()

	for _, name := range workflowNames {
		sb.WriteString(wbt[name].markdownRow(name))
	}

	return sb.String()
}

// sortWorkflowNames returns a sorted slice of workflow names.
func (wbt WorkflowBillableTime) sortWorkflowNames() []string {
	var workflowNames []string
	for name := range wbt {
		workflowNames = append(workflowNames, name)
	}
	sort.Strings(workflowNames)
	return workflowNames
}

// return string to generate markdown row for each environment
func (e EnvBillableTime) markdownRow(title string) string {
	return fmt.Sprintf("| %s | %d | %d | %d |\n", title, e.Ubuntu, e.Windows, e.Macos)
}

// return bold string to generate markdown row for each environment
func (e EnvBillableTime) markdownBoldRow(title string) string {
	return fmt.Sprintf("| **%s** | **%d** | **%d** | **%d** |\n", title, e.Ubuntu, e.Windows, e.Macos)
}

// CreateReport retrieves billable time for workflows and dumps as markdown
func CreateReport(repository string) error {
	owner, repo, err := getOwnerAndRepo(repository)
	if err != nil {
		return err
	}

	client := newGitHubClient()
	workflows, err := getWorkflows(client, owner, repo)
	if err != nil {
		return err
	}

	wbt, err := generateWorkflowBillableTime(client, owner, repo, workflows)
	if err != nil {
		return err
	}
	err = appendToFile(getOutputPath(), wbt.generateMarkdownText())
	if err != nil {
		return err
	}

	return nil
}

// newGitHubClient returns a new GitHub API client with optional authentication token
func newGitHubClient() *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return github.NewClient(nil)
	}
	return github.NewClient(nil).WithAuthToken(token)
}

// getWorkflows retrieves a list of workflows for the specified repository
func getWorkflows(client *github.Client, owner, repo string) ([]*github.Workflow, error) {
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

// generateWorkflowBillableTime generates a WorkflowBillableTime map for the specified workflows
func generateWorkflowBillableTime(client *github.Client, owner, repo string, workflows []*github.Workflow) (WorkflowBillableTime, error) {
	wbt := make(WorkflowBillableTime)

	for _, workflow := range workflows {
		billMap, err := getWorkflowBillableTime(client, owner, repo, *workflow.ID)
		if err != nil {
			return nil, err
		}

		wbt[*workflow.Name] = EnvBillableTime{
			Ubuntu:  getMinutesForEnv(billMap, "UBUNTU"),
			Windows: getMinutesForEnv(billMap, "WINDOWS"),
			Macos:   getMinutesForEnv(billMap, "MACOS"),
		}
	}

	return wbt, nil
}

// getWorkflowBillableTime retrieves the billable time map for a specific workflow
func getWorkflowBillableTime(client *github.Client, owner, repo string, workflowID int64) (github.WorkflowBillMap, error) {
	usage, _, err := client.Actions.GetWorkflowUsageByID(context.Background(), owner, repo, workflowID)
	if err != nil {
		return nil, err
	}

	return *usage.Billable, nil
}

// getOwnerAndRepo extracts the owner and repository name from the provided repository argument or environment variable
func getOwnerAndRepo(repo string) (string, string, error) {
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
