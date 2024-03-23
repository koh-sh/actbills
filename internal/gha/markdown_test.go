package gha

import (
	"os"
	"path/filepath"
	"testing"
)

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
