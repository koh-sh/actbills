package gha

import (
	"reflect"
	"testing"

	"github.com/google/go-github/v60/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

func Test_getOwnerAndRepo(t *testing.T) {
	type args struct {
		repo string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name:    "basic",
			args:    args{repo: "owner/repo"},
			want:    "owner",
			want1:   "repo",
			wantErr: false,
		},
		{
			name:    "with env",
			args:    args{repo: ""},
			want:    "owner",
			want1:   "repo",
			wantErr: false,
		},
		{
			name:    "without env",
			args:    args{repo: ""},
			want:    "",
			want1:   "",
			wantErr: true,
		},
		{
			name:    "wrong format",
			args:    args{repo: "foobar"},
			want:    "",
			want1:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "with env" {
				t.Setenv("GITHUB_REPOSITORY", "owner/repo")
			}
			got, got1, err := getOwnerAndRepo(tt.args.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOwnerAndRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getOwnerAndRepo() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getOwnerAndRepo() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_newGitHubClient(t *testing.T) {
	tests := []struct {
		name string
		want *github.Client
	}{
		{
			name: "basic",
			want: github.NewClient(nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newGitHubClient(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newGitHubClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getMinutesForEnv(t *testing.T) {
	u := int64(7200000)
	w := int64(3000000)
	billMap := github.WorkflowBillMap{
		"UBUNTU": &github.WorkflowBill{
			TotalMS: &u,
		},
		"WINDOWS": &github.WorkflowBill{
			TotalMS: &w,
		},
	}
	type args struct {
		billMap github.WorkflowBillMap
		env     string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "basic",
			args: args{billMap: billMap, env: "UBUNTU"},
			want: 120,
		},
		{
			name: "no bills",
			args: args{billMap: billMap, env: "MACOS"},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMinutesForEnv(tt.args.billMap, tt.args.env); got != tt.want {
				t.Errorf("getMinutesForEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getWorkflows(t *testing.T) {
	type args struct {
		client *github.Client
		owner  string
		repo   string
	}
	tests := []struct {
		name    string
		args    args
		want    []*github.Workflow
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				client: mockClientForListWorkflows("default"),
				owner:  "owner",
				repo:   "repo",
			},
			want: []*github.Workflow{
				{
					Name: github.String("workflow1"),
					ID:   github.Int64(123),
				},
				{
					Name: github.String("workflow2"),
					ID:   github.Int64(124),
				},
			},
			wantErr: false,
		},
		{
			name: "pages",
			args: args{
				client: mockClientForListWorkflows("pages"),
				owner:  "owner",
				repo:   "repo",
			},
			want: []*github.Workflow{
				{
					Name: github.String("workflow1"),
					ID:   github.Int64(123),
				},
				{
					Name: github.String("workflow2"),
					ID:   github.Int64(124),
				},
				{
					Name: github.String("workflow3"),
					ID:   github.Int64(125),
				},
			},
			wantErr: false,
		},
		{
			name: "empty",
			args: args{
				client: mockClientForListWorkflows("empty"),
				owner:  "owner",
				repo:   "repo",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "ratelimit",
			args: args{
				client: mockClientForListWorkflows("ratelimit"),
				owner:  "owner",
				repo:   "repo",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getWorkflows(tt.args.client, tt.args.owner, tt.args.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("getWorkflows() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getWorkflows() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateWorkflowBillableTime(t *testing.T) {
	type args struct {
		client    *github.Client
		owner     string
		repo      string
		workflows []*github.Workflow
	}
	tests := []struct {
		name    string
		args    args
		want    WorkflowBillableTime
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				client: mockClientForWorkflowUsage("basic"),
				owner:  "owner",
				repo:   "repo",
				workflows: []*github.Workflow{
					{
						Name: github.String("workflow1"),
						ID:   github.Int64(123),
					},
				},
			},
			want: WorkflowBillableTime{
				"workflow1": EnvBillableTime{
					Ubuntu:  1,
					Windows: 10,
					Macos:   0,
				},
			},
			wantErr: false,
		},
		{
			name: "ratelimit",
			args: args{
				client: mockClientForWorkflowUsage("ratelimit"),
				owner:  "owner",
				repo:   "repo",
				workflows: []*github.Workflow{
					{
						Name: github.String("workflow1"),
						ID:   github.Int64(123),
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateWorkflowBillableTime(tt.args.client, tt.args.owner, tt.args.repo, tt.args.workflows)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateWorkflowBillableTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateWorkflowBillableTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

// return mock GitHub Client for List Workflows
func mockClientForListWorkflows(ptn string) *github.Client {
	switch ptn {
	case "ratelimit":
		return github.NewClient(mock.NewMockedHTTPClient(
			mock.WithRequestMatch(
				mock.GetReposActionsWorkflowsByOwnerByRepo,
				github.Workflows{
					Workflows: []*github.Workflow{},
				},
			),
			mock.WithRateLimit(0, 0),
		),
		)
	case "empty":
		return github.NewClient(mock.NewMockedHTTPClient(
			mock.WithRequestMatch(
				mock.GetReposActionsWorkflowsByOwnerByRepo,
				github.Workflows{
					Workflows: []*github.Workflow{},
				},
			),
		),
		)
	case "pages":
		return github.NewClient(mock.NewMockedHTTPClient(
			mock.WithRequestMatchPages(
				mock.GetReposActionsWorkflowsByOwnerByRepo,
				github.Workflows{
					Workflows: []*github.Workflow{
						{
							Name: github.String("workflow1"),
							ID:   github.Int64(123),
						},
						{
							Name: github.String("workflow2"),
							ID:   github.Int64(124),
						},
					},
				},
				github.Workflows{
					Workflows: []*github.Workflow{
						{
							Name: github.String("workflow3"),
							ID:   github.Int64(125),
						},
					},
				},
			),
		))

	default:
		return github.NewClient(mock.NewMockedHTTPClient(
			mock.WithRequestMatch(
				mock.GetReposActionsWorkflowsByOwnerByRepo,
				github.Workflows{
					Workflows: []*github.Workflow{
						{
							Name: github.String("workflow1"),
							ID:   github.Int64(123),
						},
						{
							Name: github.String("workflow2"),
							ID:   github.Int64(124),
						},
					},
				},
			),
		))
	}
}

// return mock GitHub Client for List Workflows
func mockClientForWorkflowUsage(ptn string) *github.Client {
	switch ptn {
	case "ratelimit":
		return github.NewClient(mock.NewMockedHTTPClient(
			mock.WithRequestMatch(
				mock.GetReposActionsWorkflowsTimingByOwnerByRepoByWorkflowId,
				github.WorkflowBillMap{},
			),
			mock.WithRateLimit(0, 0),
		),
		)
	default:
		u := int64(60000)
		w := int64(600000)
		return github.NewClient(mock.NewMockedHTTPClient(
			mock.WithRequestMatch(
				mock.GetReposActionsWorkflowsTimingByOwnerByRepoByWorkflowId,
				github.WorkflowUsage{
					Billable: &github.WorkflowBillMap{
						"UBUNTU": &github.WorkflowBill{
							TotalMS: &u,
						},
						"WINDOWS": &github.WorkflowBill{
							TotalMS: &w,
						},
					},
				},
			),
		))
	}
}

func TestWorkflowBillableTime_generateMarkdownText(t *testing.T) {
	workflowBillableTime := WorkflowBillableTime{
		"Workflow2": EnvBillableTime{
			Ubuntu:  180,
			Windows: 30,
		},
		"Workflow1": EnvBillableTime{
			Ubuntu:  120,
			Windows: 90,
			Macos:   60,
		},
	}
	want := `# Billable time for workflows in this billable cycle

| Workflow | Ubuntu (min) | Windows (min) | Macos (min) |
| --- | --- | --- | --- |
| Workflow1 | 120 | 90 | 60 |
| Workflow2 | 180 | 30 | 0 |
| **Total** | **300** | **120** | **60** |

Please note the following:

- This list shows the execution time for each Workflow at the time this Action was executed.
- Workflows that have been deleted at the time of execution will not be listed.
- Execution times using Larger runners are not included in the aggregation.
`
	tests := []struct {
		name string
		w    WorkflowBillableTime
		want string
	}{
		{
			name: "basic",
			w:    workflowBillableTime,
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.w.generateMarkdownText(); got != tt.want {
				t.Errorf("WorkflowBillableTime.generateMarkdownText() = %v, want %v", got, tt.want)
			}
		})
	}
}
