package gha

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_generateMarkdownText(t *testing.T) {
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
	want := `# Billable time for each Workflows in this billable cycle

| Workflow | Ubuntu (min) | Windows (min) | Macos (min) |
| --- | --- | --- | --- |
| Workflow1 | 120 | 90 | 60 |
| Workflow2 | 180 | 30 | 0 |

Please note the following:

- This list shows the execution time for each Workflow at the time this Action was executed.
- Workflows that have been deleted at the time of execution will not be listed.
- Execution times using Larger runners are not included in the aggregation.
`
	type args struct {
		wbt WorkflowBillableTime
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic",
			args: args{wbt: workflowBillableTime},
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateMarkdownText(tt.args.wbt); got != tt.want {
				t.Errorf("generateMarkdownText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getOutputPath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "with env",
			want: "/path/to/out",
		},
		{
			name: "without env",
			want: "/dev/stdout",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "with env" {
				t.Setenv("GITHUB_STEP_SUMMARY", "/path/to/out")
			}
			if got := getOutputPath(); got != tt.want {
				t.Errorf("getOutputPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_appendToFile(t *testing.T) {
	tempDir := t.TempDir()

	type args struct {
		filePath string
		content  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Append to new file",
			args: args{
				filePath: filepath.Join(tempDir, "test1.txt"),
				content:  "Hello, ",
			},
			wantErr: false,
		},
		{
			name: "Append to existing file",
			args: args{
				filePath: filepath.Join(tempDir, "test2.txt"),
				content:  "World!",
			},
			wantErr: false,
		},
		{
			name: "Append to file in nonexistent directory",
			args: args{
				filePath: "/path/to/nonexistent/directory/file.txt",
				content:  "Test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var expectedContent string
			if tt.name == "Append to existing file" {
				// For the test case that appends to an existing file, create the file beforehand
				initialContent := "Hello, "
				err := os.WriteFile(tt.args.filePath, []byte(initialContent), 0o644)
				if err != nil {
					t.Fatalf("Failed to create file: %v", err)
				}
				expectedContent = initialContent + tt.args.content
			} else {
				expectedContent = tt.args.content
			}

			err := appendToFile(tt.args.filePath, tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("appendToFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// If no error is expected, verify the file content
				content, err := os.ReadFile(tt.args.filePath)
				if err != nil {
					t.Errorf("Failed to read file: %v", err)
					return
				}
				if string(content) != expectedContent {
					t.Errorf("File content mismatch. Got %q, want %q", string(content), expectedContent)
				}
			}
		})
	}
}
