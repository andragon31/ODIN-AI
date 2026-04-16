package guardian

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/odin-ai/odin/pkg/logger"
)

// InstallHook installs the pre-commit hook
func InstallHook() error {
	// Get git hooks path
	hooksPath, err := getGitHooksPath()
	if err != nil {
		return fmt.Errorf("failed to get git hooks path: %w", err)
	}

	// Ensure hooks directory exists
	if err := os.MkdirAll(hooksPath, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	// Create pre-commit hook content
	hookContent := getPreCommitHookScript()

	// Write hook file
	hookPath := filepath.Join(hooksPath, "pre-commit")
	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return fmt.Errorf("failed to write pre-commit hook: %w", err)
	}

	logger.Info("Pre-commit hook installed", "path", hookPath)
	return nil
}

// UninstallHook removes the pre-commit hook
func UninstallHook() error {
	hooksPath, err := getGitHooksPath()
	if err != nil {
		return fmt.Errorf("failed to get git hooks path: %w", err)
	}

	hookPath := filepath.Join(hooksPath, "pre-commit")
	if _, err := os.Stat(hookPath); err == nil {
		if err := os.Remove(hookPath); err != nil {
			return fmt.Errorf("failed to remove pre-commit hook: %w", err)
		}
		logger.Info("Pre-commit hook removed")
	}

	return nil
}

// IsHookInstalled checks if the pre-commit hook is installed
func IsHookInstalled() bool {
	hooksPath, err := getGitHooksPath()
	if err != nil {
		return false
	}

	hookPath := filepath.Join(hooksPath, "pre-commit")
	_, err = os.Stat(hookPath)
	return err == nil
}

// getGitHooksPath returns the path to the .git/hooks directory
func getGitHooksPath() (string, error) {
	// Check if we're in a git repository
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return "", fmt.Errorf("not in a git repository")
	}

	// Get absolute path to .git/hooks
	absPath, err := filepath.Abs(".git/hooks")
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}

// getPreCommitHookScript returns the pre-commit hook script content
func getPreCommitHookScript() string {
	return `#!/bin/sh
# ODIN Heimdall Pre-Commit Hook
# This hook runs security checks on staged files before commit

# Exit on any error
set -e

# Get the ODIN binary path
ODIN="${ODIN:-odin}"

# Check if heimdall command exists
if ! $ODIN heimdall --help > /dev/null 2>&1; then
    echo "Heimdall not found. Skipping security checks."
    exit 0
fi

# Run heimdall check on staged files
echo "Running Heimdall security checks..."

# Get staged files
STAGED_FILES=$(git diff --cached --name-only --diff-filter=ACM)

if [ -z "$STAGED_FILES" ]; then
    echo "No staged files to check."
    exit 0
fi

# Run heimdall check
CHECK_RESULT=0
$ODIN heimdall check $STAGED_FILES || CHECK_RESULT=$?

# Exit codes:
# 0 = passed (no issues)
# 1 = failed (issues found, but not blocking)
# 2 = error

case $CHECK_RESULT in
    0)
        echo "Security check passed!"
        exit 0
        ;;
    1)
        # Issues found but not critical - warn but allow commit
        echo "Warning: Security issues found. Review above."
        exit 1
        ;;
    2)
        # Critical issues - block commit
        echo "ERROR: Critical security issues found. Commit blocked."
        echo "Fix the issues above or use --no-verify to skip."
        exit 1
        ;;
    *)
        # Unexpected error - warn but allow commit
        echo "Warning: Security check encountered an error."
        exit 0
        ;;
esac
`
}

// getPreCommitHookScriptWindows returns the Windows pre-commit hook script
func getPreCommitHookScriptWindows() string {
	return `@echo off
REM ODIN Heimdall Pre-Commit Hook for Windows
REM This hook runs security checks on staged files before commit

REM Get the ODIN binary path
set ODIN=odin

REM Check if heimdall command exists
where /Q %ODIN% >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Heimdall not found. Skipping security checks.
    exit /b 0
)

REM Run heimdall check on staged files
echo Running Heimdall security checks...

REM Get staged files
for /f "delims=" %%i in ('git diff --cached --name-only --diff-filter=ACM') do set STAGED_FILES=!STAGED_FILES! %%i

if "!STAGED_FILES!"=="" (
    echo No staged files to check.
    exit /b 0
)

REM Run heimdall check
call %ODIN% heimdall check %STAGED_FILES%
set CHECK_RESULT=%ERRORLEVEL%

if %CHECK_RESULT% equ 0 (
    echo Security check passed!
    exit /b 0
) else if %CHECK_RESULT% equ 1 (
    echo Warning: Security issues found. Review above.
    exit /b 0
) else (
    echo ERROR: Critical security issues found. Commit blocked.
    exit /b 1
)
`
}

// RunHookManual runs the pre-commit hook manually (for testing)
func RunHookManual(ctx context.Context) error {
	// Get staged files
	stagedFiles, err := getStagedFiles()
	if err != nil {
		return fmt.Errorf("failed to get staged files: %w", err)
	}

	if len(stagedFiles) == 0 {
		fmt.Println("No staged files to check.")
		return nil
	}

	// Build odin heimdall check command
	args := append([]string{"heimdall", "check"}, stagedFiles...)
	cmd := exec.CommandContext(ctx, "odin", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// ValidateHookScript validates the pre-commit hook script
func ValidateHookScript() error {
	hooksPath, err := getGitHooksPath()
	if err != nil {
		return err
	}

	hookPath := filepath.Join(hooksPath, "pre-commit")
	content, err := os.ReadFile(hookPath)
	if err != nil {
		return fmt.Errorf("pre-commit hook not found: %w", err)
	}

	// Check if hook contains Heimdall reference
	if !strings.Contains(string(content), "Heimdall") && !strings.Contains(string(content), "heimdall") {
		return fmt.Errorf("pre-commit hook does not reference Heimdall")
	}

	return nil
}
