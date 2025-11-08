package tty

import (
	"os"
	"strings"
	"testing"
)

func TestIsCI(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name:     "no CI vars",
			envVars:  map[string]string{},
			expected: false,
		},
		{
			name:     "generic CI var",
			envVars:  map[string]string{"CI": "true"},
			expected: true,
		},
		{
			name:     "GitHub Actions",
			envVars:  map[string]string{"GITHUB_ACTIONS": "true"},
			expected: true,
		},
		{
			name:     "GitLab CI",
			envVars:  map[string]string{"GITLAB_CI": "true"},
			expected: true,
		},
		{
			name:     "CircleCI",
			envVars:  map[string]string{"CIRCLECI": "true"},
			expected: true,
		},
		{
			name:     "multiple CI vars",
			envVars:  map[string]string{"CI": "true", "TRAVIS": "true"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciEnvVars := []string{
				"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS",
				"GITLAB_CI", "CIRCLECI", "TRAVIS", "JENKINS_URL",
				"BUILDKITE", "DRONE", "TEAMCITY_VERSION",
			}
			for _, v := range ciEnvVars {
				os.Unsetenv(v)
			}

			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			result := IsCI()
			if result != tt.expected {
				t.Errorf("IsCI() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetCIName(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		expected string
	}{
		{
			name:     "GitHub Actions",
			envVar:   "GITHUB_ACTIONS",
			expected: "GitHub Actions",
		},
		{
			name:     "GitLab CI",
			envVar:   "GITLAB_CI",
			expected: "GitLab CI",
		},
		{
			name:     "CircleCI",
			envVar:   "CIRCLECI",
			expected: "CircleCI",
		},
		{
			name:     "Travis CI",
			envVar:   "TRAVIS",
			expected: "Travis CI",
		},
		{
			name:     "Jenkins",
			envVar:   "JENKINS_URL",
			expected: "Jenkins",
		},
		{
			name:     "Buildkite",
			envVar:   "BUILDKITE",
			expected: "Buildkite",
		},
		{
			name:     "Drone CI",
			envVar:   "DRONE",
			expected: "Drone CI",
		},
		{
			name:     "TeamCity",
			envVar:   "TEAMCITY_VERSION",
			expected: "TeamCity",
		},
		{
			name:     "Generic CI",
			envVar:   "CI",
			expected: "CI",
		},
		{
			name:     "No CI",
			envVar:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciEnvVars := []string{
				"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS",
				"GITLAB_CI", "CIRCLECI", "TRAVIS", "JENKINS_URL",
				"BUILDKITE", "DRONE", "TEAMCITY_VERSION",
			}
			for _, v := range ciEnvVars {
				os.Unsetenv(v)
			}

			if tt.envVar != "" {
				os.Setenv(tt.envVar, "true")
			}

			defer func() {
				if tt.envVar != "" {
					os.Unsetenv(tt.envVar)
				}
			}()

			result := GetCIName()
			if result != tt.expected {
				t.Errorf("GetCIName() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestErrorInteractiveRequired(t *testing.T) {
	tests := []struct {
		name         string
		commandName  string
		alternatives []string
		ciEnv        string
		wantContains []string
	}{
		{
			name:        "basic error",
			commandName: "review",
			wantContains: []string{
				"command 'review' requires an interactive terminal",
			},
		},
		{
			name:        "with alternatives",
			commandName: "review",
			alternatives: []string{
				"Use 'storm unreleased list' to view entries",
				"Use 'storm unreleased list --json' for JSON output",
			},
			wantContains: []string{
				"command 'review' requires an interactive terminal",
				"Alternatives:",
				"Use 'storm unreleased list' to view entries",
				"Use 'storm unreleased list --json' for JSON output",
			},
		},
		{
			name:        "CI environment",
			commandName: "diff",
			ciEnv:       "GITHUB_ACTIONS",
			wantContains: []string{
				"command 'diff' requires an interactive terminal",
				"detected GitHub Actions environment",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciEnvVars := []string{
				"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS",
				"GITLAB_CI", "CIRCLECI", "TRAVIS", "JENKINS_URL",
				"BUILDKITE", "DRONE", "TEAMCITY_VERSION",
			}
			for _, v := range ciEnvVars {
				os.Unsetenv(v)
			}

			if tt.ciEnv != "" {
				os.Setenv(tt.ciEnv, "true")
				defer os.Unsetenv(tt.ciEnv)
			}

			err := ErrorInteractiveRequired(tt.commandName, tt.alternatives)
			if err == nil {
				t.Fatal("ErrorInteractiveRequired() returned nil, expected error")
			}

			errMsg := err.Error()
			for _, want := range tt.wantContains {
				if !strings.Contains(errMsg, want) {
					t.Errorf("ErrorInteractiveRequired() error message missing %q\nGot: %s", want, errMsg)
				}
			}
		})
	}
}

func TestErrorInteractiveFlag(t *testing.T) {
	tests := []struct {
		name         string
		flagName     string
		ciEnv        string
		wantContains []string
	}{
		{
			name:     "basic error",
			flagName: "--interactive",
			wantContains: []string{
				"flag '--interactive' requires an interactive terminal",
			},
		},
		{
			name:     "CI environment",
			flagName: "--interactive",
			ciEnv:    "GITLAB_CI",
			wantContains: []string{
				"flag '--interactive' requires an interactive terminal",
				"detected GitLab CI environment",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciEnvVars := []string{
				"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS",
				"GITLAB_CI", "CIRCLECI", "TRAVIS", "JENKINS_URL",
				"BUILDKITE", "DRONE", "TEAMCITY_VERSION",
			}
			for _, v := range ciEnvVars {
				os.Unsetenv(v)
			}

			if tt.ciEnv != "" {
				os.Setenv(tt.ciEnv, "true")
				defer os.Unsetenv(tt.ciEnv)
			}

			err := ErrorInteractiveFlag(tt.flagName)
			if err == nil {
				t.Fatal("ErrorInteractiveFlag() returned nil, expected error")
			}

			errMsg := err.Error()
			for _, want := range tt.wantContains {
				if !strings.Contains(errMsg, want) {
					t.Errorf("ErrorInteractiveFlag() error message missing %q\nGot: %s", want, errMsg)
				}
			}
		})
	}
}
