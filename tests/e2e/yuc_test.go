package test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build the binary
	tmpDir, err := os.MkdirTemp("", "yuc-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binaryName := "yuc"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath = filepath.Join(tmpDir, binaryName)

	// Get path to project root
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "../..")

	cmd := exec.Command("go", "build", "-o", binaryPath, "main.go")
	cmd.Dir = projectRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n%s\n", err, output)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func runYuc(args ...string) (string, string, int, error) {
	cmd := exec.Command(binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return "", "", 0, err
		}
	}
	return stdout.String(), stderr.String(), exitCode, nil
}

func TestYucCommandsAndFlags(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedExit   int
		expectStdout   string
		expectStderr   string
		containsStdout []string
		containsStderr []string
	}{
		{
			name:           "help flag",
			args:           []string{"--help"},
			expectedExit:   0,
			containsStdout: []string{"USAGE:", "FLAGS:", "EXAMPLES:"},
		},
		{
			name:           "short help flag",
			args:           []string{"-h"},
			expectedExit:   0,
			containsStdout: []string{"USAGE:", "FLAGS:", "EXAMPLES:"},
		},
		{
			name:           "version flag",
			args:           []string{"--version"},
			expectedExit:   0,
			containsStdout: []string{"yuc"},
		},
		{
			name:           "short version flag",
			args:           []string{"-v"},
			expectedExit:   0,
			containsStdout: []string{"yuc"},
		},
		{
			name:           "list categories flag",
			args:           []string{"--list-categories"},
			expectedExit:   0,
			containsStdout: []string{"yuc built-in risk categories:", "BIDI_CONTROL", "ZERO_WIDTH", "HOMOGLYPH"},
		},
		{
			name:           "no args (shows help)",
			args:           []string{},
			expectedExit:   2,
			containsStderr: []string{"USAGE:", "FLAGS:"},
		},
		{
			name:           "safe file",
			args:           []string{"testdata/001-safe.yaml"},
			expectedExit:   0,
			containsStdout: []string{"0 errors", "0 warnings"},
		},
		{
			name:           "file with issue (warning)",
			args:           []string{"testdata/002-issue.yaml"},
			expectedExit:   1,
			containsStdout: []string{"WARN", "ZERO_WIDTH", "0 errors", "1 warning"},
		},
		{
			name:           "file with error",
			args:           []string{"testdata/003-error.yaml"},
			expectedExit:   1,
			containsStdout: []string{"ERROR", "BIDI_CONTROL", "1 error", "0 warnings"},
		},
		{
			name:           "file with nbsp",
			args:           []string{"testdata/004-nbsp-error.yaml"},
			expectedExit:   1,
			containsStdout: []string{"WARN", "OVERLONG_SPACE", "0 error", "2 warnings"},
		},
		{
			name:           "short config flag - disable category",
			args:           []string{"-c", "testdata/config.yaml", "testdata/002-issue.yaml"},
			expectedExit:   0,
			containsStdout: []string{"0 errors", "0 warnings"},
		},
		{
			name:           "multiple files",
			args:           []string{"testdata/001-safe.yaml", "testdata/002-issue.yaml"},
			expectedExit:   1,
			containsStdout: []string{"yuc: scanning 2 files", "1 clean", "1 with issues", "0 errors", "1 warning"},
		},
		{
			name:           "invalid config path",
			args:           []string{"--config", "nonexistent.yaml", "testdata/001-safe.yaml"},
			expectedExit:   3,
			containsStderr: []string{"config error"},
		},
		{
			name:           "nonexistent input file",
			args:           []string{"nonexistent.yaml"},
			expectedExit:   2,
			containsStderr: []string{"cannot open \"nonexistent.yaml\""},
		},
		{
			name:           "no color flag",
			args:           []string{"--no-color", "testdata/002-issue.yaml"},
			expectedExit:   1,
			containsStdout: []string{"0 errors", "1 warning"},
		},
		{
			name:           "config flag - severity override",
			args:           []string{"--config", "testdata/config-severity.yaml", "testdata/002-issue.yaml"},
			expectedExit:   1,
			containsStdout: []string{"ERROR", "1 error", "0 warnings"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, exitCode, err := runYuc(tc.args...)
			if err != nil {
				t.Fatalf("failed to run yuc: %v", err)
			}

			if exitCode != tc.expectedExit {
				t.Errorf("expected exit code %d, got %d. Stdout: %s, Stderr: %s", tc.expectedExit, exitCode, stdout, stderr)
			}

			for _, s := range tc.containsStdout {
				if !strings.Contains(stdout, s) {
					t.Errorf("expected stdout to contain %q, but it didn't. Stdout: %s", s, stdout)
				}
			}

			for _, s := range tc.containsStderr {
				if !strings.Contains(stderr, s) {
					t.Errorf("expected stderr to contain %q, but it didn't. Stderr: %s", s, stderr)
				}
			}

			if strings.Contains(tc.name, "no color flag") {
				if strings.Contains(stdout, "\033[") {
					t.Errorf("expected stdout to not contain ANSI escape codes with --no-color flag. Stdout: %q", stdout)
				}
			}
		})
	}
}
