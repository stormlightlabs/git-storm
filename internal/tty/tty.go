// package tty provides utilities for detecting terminal (TTY) availability and
// generating appropriate fallback behavior for non-interactive environments.
package tty

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

// IsTTY checks if the given file descriptor is a terminal.
func IsTTY(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}

// IsInteractive checks if both stdin and stdout are connected to a terminal.
// This is the primary check for determining if TUI applications can run.
func IsInteractive() bool {
	return IsTTY(os.Stdin.Fd()) && IsTTY(os.Stdout.Fd())
}

// IsCI detects if the current environment is a CI system by checking for common
// CI environment variables.
func IsCI() bool {
	ciEnvVars := []string{
		"CI", // Generic CI indicator
		"CONTINUOUS_INTEGRATION",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"CIRCLECI",
		"TRAVIS",
		"JENKINS_URL",
		"BUILDKITE",
		"DRONE",
		"TEAMCITY_VERSION",
	}

	for _, envVar := range ciEnvVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

// GetCIName attempts to identify the specific CI system being used.
func GetCIName() string {
	ciMap := map[string]string{
		"GITHUB_ACTIONS":   "GitHub Actions",
		"GITLAB_CI":        "GitLab CI",
		"CIRCLECI":         "CircleCI",
		"TRAVIS":           "Travis CI",
		"JENKINS_URL":      "Jenkins",
		"BUILDKITE":        "Buildkite",
		"DRONE":            "Drone CI",
		"TEAMCITY_VERSION": "TeamCity",
	}

	for envVar, name := range ciMap {
		if os.Getenv(envVar) != "" {
			return name
		}
	}

	if IsCI() {
		return "CI"
	}

	return ""
}

// ErrorInteractiveRequired returns a formatted error message indicating that the
// command requires an interactive terminal, with suggestions for alternatives.
func ErrorInteractiveRequired(commandName string, alternatives []string) error {
	msg := fmt.Sprintf("command '%s' requires an interactive terminal", commandName)

	if IsCI() {
		ciName := GetCIName()
		msg += fmt.Sprintf(" (detected %s environment)", ciName)
	} else {
		msg += " (stdin is not a TTY)"
	}

	if len(alternatives) > 0 {
		msg += "\n\nAlternatives:"
		for _, alt := range alternatives {
			msg += fmt.Sprintf("\n  - %s", alt)
		}
	}

	return errors.New(msg)
}

// ErrorInteractiveFlag returns a formatted error message indicating that an
// interactive flag cannot be used in a non-TTY environment.
func ErrorInteractiveFlag(flagName string) error {
	msg := fmt.Sprintf("flag '%s' requires an interactive terminal", flagName)

	if IsCI() {
		ciName := GetCIName()
		msg += fmt.Sprintf(" (detected %s environment)", ciName)
	} else {
		msg += " (stdin is not a TTY)"
	}

	return errors.New(msg)
}
