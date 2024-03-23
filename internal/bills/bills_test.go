package bills

import (
	"reflect"
	"testing"

	"github.com/google/go-github/v60/github"
)

func Test_generateWorkflowBillableTimes(t *testing.T) {
	type args struct {
		client    *github.Client
		owner     string
		repo      string
		workflows []*github.Workflow
	}
	tests := []struct {
		name    string
		args    args
		want    WorkflowBillableTimes
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
			want: WorkflowBillableTimes{
				"workflow1": WorkflowBillableTime{
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
			got, err := generateWorkflowBillableTimes(tt.args.client, tt.args.owner, tt.args.repo, tt.args.workflows)
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

func TestWorkflowBillableTimes_generateMarkdownReport(t *testing.T) {
	workflowBillableTimes := WorkflowBillableTimes{
		"Workflow2": WorkflowBillableTime{
			Ubuntu:  180,
			Windows: 30,
		},
		"Workflow1": WorkflowBillableTime{
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
		w    WorkflowBillableTimes
		want string
	}{
		{
			name: "basic",
			w:    workflowBillableTimes,
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.w.generateMarkdownReport(); got != tt.want {
				t.Errorf("WorkflowBillableTime.generateMarkdownText() = %v, want %v", got, tt.want)
			}
		})
	}
}
