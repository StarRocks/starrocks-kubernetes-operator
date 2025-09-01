package pod

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Version
		hasError bool
	}{
		{
			name:     "valid standard version",
			input:    "3.3.17",
			expected: Version{Major: 3, Minor: 3, Patch: 17},
			hasError: false,
		},
		{
			name:     "valid version with suffix",
			input:    "3.3.17-rc1",
			expected: Version{Major: 3, Minor: 3, Patch: 17},
			hasError: false,
		},
		{
			name:     "valid version with complex suffix",
			input:    "3.4.6-alpha.1.beta",
			expected: Version{Major: 3, Minor: 4, Patch: 6},
			hasError: false,
		},
		{
			name:     "invalid format - missing patch",
			input:    "3.3",
			expected: Version{},
			hasError: true,
		},
		{
			name:     "invalid format - too many parts",
			input:    "3.3.17.1",
			expected: Version{},
			hasError: true,
		},
		{
			name:     "invalid major version",
			input:    "a.3.17",
			expected: Version{},
			hasError: true,
		},
		{
			name:     "invalid minor version",
			input:    "3.b.17",
			expected: Version{},
			hasError: true,
		},
		{
			name:     "invalid patch version",
			input:    "3.3.c",
			expected: Version{},
			hasError: true,
		},
		{
			name:     "invalid patch version",
			input:    "3.3-latest",
			expected: Version{},
			hasError: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: Version{},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseVersion(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error for input %s, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %s: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("for input %s, expected %+v, got %+v", tt.input, tt.expected, result)
				}
			}
		})
	}
}

func TestIsLowerThanAny(t *testing.T) {
	targetVersions := []string{"3.3.17", "3.4.6", "3.5.2"}

	tests := []struct {
		name         string
		checkVersion string
		expected     bool
		hasError     bool
	}{
		{
			name:         "lower than 3.3.17",
			checkVersion: "3.3.9",
			expected:     true,
			hasError:     false,
		},
		{
			name:         "lower than 3.3.17",
			checkVersion: "3.3.9-ee",
			expected:     true,
			hasError:     false,
		},
		{
			name:         "lower than 3.3.17",
			checkVersion: "3.3.16",
			expected:     true,
			hasError:     false,
		},
		{
			name:         "lower than 3.3.17",
			checkVersion: "3.3.16-ee",
			expected:     true,
			hasError:     false,
		},
		{
			name:         "lower than 3.3.17",
			checkVersion: "3.3-latest",
			expected:     false,
			hasError:     false,
		},
		{
			name:         "equal to 3.3.17",
			checkVersion: "3.3.17",
			expected:     false,
			hasError:     false,
		},
		{
			name:         "equal to 3.3.17",
			checkVersion: "3.3.17-ee",
			expected:     false,
			hasError:     false,
		},
		{
			name:         "higher than 3.3.17",
			checkVersion: "3.3.18",
			expected:     false,
			hasError:     false,
		},
		{
			name:         "higher than 3.3.17",
			checkVersion: "3.3.18-ee",
			expected:     false,
			hasError:     false,
		},
		{
			name:         "lower than 3.4.6",
			checkVersion: "3.4.5",
			expected:     true,
			hasError:     false,
		},
		{
			name:         "lower than 3.5.2",
			checkVersion: "3.5.1",
			expected:     true,
			hasError:     false,
		},
		{
			name:         "higher than 3.5.2",
			checkVersion: "3.5.10",
			expected:     false,
			hasError:     false,
		},
		{
			name:         "higher than 3.5.2  -2",
			checkVersion: "4.0.0",
			expected:     false,
			hasError:     true,
		},
		{
			name:         "different major.minor - 3.2.x",
			checkVersion: "3.2.20",
			expected:     false,
			hasError:     true,
		},
		{
			name:         "different major.minor - 3.6.x",
			checkVersion: "3.6.1",
			expected:     false,
			hasError:     true,
		},
		{
			name:         "different major version",
			checkVersion: "4.3.16",
			expected:     false,
			hasError:     true,
		},
		{
			name:         "version with suffix - lower",
			checkVersion: "3.3.16-rc1",
			expected:     true,
			hasError:     false,
		},
		{
			name:         "version with suffix - equal",
			checkVersion: "3.3.17-beta",
			expected:     false,
			hasError:     false,
		},
		{
			name:         "version with suffix - higher",
			checkVersion: "3.3.18-alpha",
			expected:     false,
			hasError:     false,
		},
		{
			name:         "invalid check version format",
			checkVersion: "3.3",
			expected:     false,
			hasError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IsLowerThanAny(tt.checkVersion, targetVersions)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error for checkVersion %s, but got none", tt.checkVersion)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for checkVersion %s: %v", tt.checkVersion, err)
				}
				if result != tt.expected {
					t.Errorf("IsLowerThanAny(%s, %v) = %t, expected %t", tt.checkVersion, targetVersions, result, tt.expected)
				}
			}
		})
	}
}

func TestIsLowerThanAnyWithInvalidTargetVersions(t *testing.T) {
	tests := []struct {
		name           string
		checkVersion   string
		targetVersions []string
		hasError       bool
	}{
		{
			name:           "empty target versions",
			checkVersion:   "3.3.16",
			targetVersions: []string{},
			hasError:       true,
		},
		{
			name:           "all valid versions",
			checkVersion:   "3.3.16",
			targetVersions: []string{"3.3.17", "3.4.6", "3.5.2"},
			hasError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := IsLowerThanAny(tt.checkVersion, tt.targetVersions)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error for targetVersions %v, but got none", tt.targetVersions)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for targetVersions %v: %v", tt.targetVersions, err)
				}
			}
		})
	}
}

func TestIsLowerThanAnyEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		checkVersion   string
		targetVersions []string
		expected       bool
		hasError       bool
	}{
		{
			name:           "multiple matching major.minor versions",
			checkVersion:   "3.3.15",
			targetVersions: []string{"3.3.16", "3.3.17", "3.3.18"},
			expected:       true,
			hasError:       false,
		},
		{
			name:           "check version higher than all targets in same major.minor",
			checkVersion:   "3.3.20",
			targetVersions: []string{"3.3.16", "3.3.17", "3.3.18"},
			expected:       false,
			hasError:       false,
		},
		{
			name:           "single target version - lower",
			checkVersion:   "3.3.16",
			targetVersions: []string{"3.3.17"},
			expected:       true,
			hasError:       false,
		},
		{
			name:           "single target version - higher",
			checkVersion:   "3.3.18",
			targetVersions: []string{"3.3.17"},
			expected:       false,
			hasError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IsLowerThanAny(tt.checkVersion, tt.targetVersions)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("IsLowerThanAny(%s, %v) = %t, expected %t", tt.checkVersion, tt.targetVersions, result, tt.expected)
				}
			}
		})
	}
}
