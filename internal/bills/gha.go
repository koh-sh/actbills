package bills

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v60/github"
)

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
