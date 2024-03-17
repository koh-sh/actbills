package gha

import (
	"context"
	"fmt"
	"os"
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

// retrieve billable time for workflows and dump as markdown
func CreateReport(repository string) error {
	owner, repo, err := getOwnerAndRepo(repository)
	if err != nil {
		return err
	}
	// out := getOutputPath()
	client := rtnClient()
	workflows, err := getWorkflows(client, owner, repo)
	if err != nil {
		return err
	}
	wbt := WorkflowBillableTime{}
	for _, v := range workflows {
		m, err := getWorkflowBillableTime(client, owner, repo, *v.ID)
		if err != nil {
			return err
		}
		wbt[*v.Name] = EnvBillableTime{
			Ubuntu:  getTotalMinutesForEnv(m, "UBUNTU"),
			Windows: getTotalMinutesForEnv(m, "WINDOWS"),
			Macos:   getTotalMinutesForEnv(m, "MACOS"),
		}
	}
	for k, v := range wbt {
		fmt.Printf("%s, %d, %d, %d\n", k, v.Ubuntu, v.Windows, v.Macos)
	}
	return nil
}

// return client of github api
func rtnClient() *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return github.NewClient(nil)
	}
	return github.NewClient(nil).WithAuthToken(token)
}

// get list of workflows for repository
func getWorkflows(client *github.Client, owner, repo string) ([]*github.Workflow, error) {
	opts := &github.ListOptions{
		PerPage: 100,
	}

	var allWorkflows []*github.Workflow
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

// get billable timemap for a workflow
func getWorkflowBillableTime(client *github.Client, owner, repo string, workflowID int64) (github.WorkflowBillMap, error) {
	usage, _, err := client.Actions.GetWorkflowUsageByID(context.Background(), owner, repo, workflowID)
	if err != nil {
		return nil, err
	}

	return *usage.Billable, nil
}

// getOwnerAndRepo extracts the owner and repository name from the provided `repo` argument
// or the GITHUB_REPOSITORY environment variable based on the following conditions:
//   - If the `repo` argument is not an empty string, it splits the value by the slash (/) and
//     returns the owner and repo values.
//   - If the `repo` argument is an empty string and the environment variable GITHUB_REPOSITORY is set,
//     it splits the value of GITHUB_REPOSITORY by the slash (/) and returns the owner and repo values.
//   - If neither of the above conditions are met, it returns an error.
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

// GetTotalMinutesPerEnvironment takes a WorkflowBillMap and returns a map containing the total time in minutes
// for each of the specified environments: UBUNTU, MACOS, and WINDOWS. If a key doesn't exist in the input map,
// the corresponding value in the result map will be set to 0.
func getTotalMinutesForEnv(billMap github.WorkflowBillMap, env string) int64 {
	bill, ok := billMap[env]
	if !ok {
		return 0
	}

	return bill.GetTotalMS() / 60000 // convert milliseconds to minutes
}
