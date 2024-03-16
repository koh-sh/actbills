package gha

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v60/github"
)

// retrieve billable time for workflows and dump as markdown
func CreateReport(repository string) error {
	owner, repo, err := getOwnerAndRepo(repository)
	if err != nil {
		return err
	}
	out := getOutputPath()
	client := rtnClient()
	workflows, err := getWorkflows(client, owner, repo)
	if err != nil {
		return err
	}
	for _, v := range workflows {
		// WIP
		err = appendToFile(out, *v.Name)
		if err != nil {
			return err
		}
		m, _ := getWorkflowBillableTime(client, owner, repo, *v.ID)
		v, ok := m["UBUNTU"]
		if ok {
			fmt.Println(v.GetTotalMS())
		}
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

	billableTime := usage.Billable

	return *billableTime, nil
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
