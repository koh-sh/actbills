package gha

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/google/go-github/v60/github"
)

// EnvBillableTime represents the total billable time for each environment in a workflow
type EnvBillableTime struct {
	Ubuntu  int64 // Total billable time for the Ubuntu environment (in minutes)
	Windows int64 // Total billable time for the Windows environment (in minutes)
	Macos   int64 // Total billable time for the Mac environment (in minutes)
}

// WorkflowBillableTime represents a map of workflow names to their corresponding EnvBillableTime
type WorkflowBillableTime map[string]EnvBillableTime

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

	wbt := generateWorkflowBillableTime(client, owner, repo, workflows)
	printWorkflowBillableTime(wbt)

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
func generateWorkflowBillableTime(client *github.Client, owner, repo string, workflows []*github.Workflow) WorkflowBillableTime {
	wbt := make(WorkflowBillableTime)

	for _, workflow := range workflows {
		billMap, err := getWorkflowBillableTime(client, owner, repo, *workflow.ID)
		if err != nil {
			continue
		}

		wbt[*workflow.Name] = EnvBillableTime{
			Ubuntu:  getMinutesForEnv(billMap, "UBUNTU"),
			Windows: getMinutesForEnv(billMap, "WINDOWS"),
			Macos:   getMinutesForEnv(billMap, "MACOS"),
		}
	}

	return wbt
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

// printWorkflowBillableTime prints the WorkflowBillableTime data in a sorted order
func printWorkflowBillableTime(wbt WorkflowBillableTime) {
	// Sort the workflow names
	var workflowNames []string
	for name := range wbt {
		workflowNames = append(workflowNames, name)
	}
	sort.Strings(workflowNames)

	// Print the data in the sorted order
	for _, name := range workflowNames {
		envTime := wbt[name]
		fmt.Printf("%s, %d, %d, %d\n", name, envTime.Ubuntu, envTime.Windows, envTime.Macos)
	}
}
