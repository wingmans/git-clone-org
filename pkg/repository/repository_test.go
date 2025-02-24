package repository_test

import (
	"testing"

	r "wingmen.io/git-clone-all/pkg/repository"
)

func TestLoad(t *testing.T) {
	filename := "/Users/paul/wingman/github.com/git-clone-org/testdata/.gitexclude"
	result := r.LoadExcludedRepos(filename)

	t.Logf("loaded %v exceptions", len(result))

	if len(result) != 2 {
		t.Errorf("expected 2 got %v", len(result))
	}
	t.Log("success")
}

func TestLoadExcludedRepos(t *testing.T) {

	tests := []struct {
		name     string
		filename string
		expected []string
	}{
		{
			name:     "Valid file with repos",
			filename: "/Users/paul/wingman/github.com/git-clone-org/testdata/.gitexclude",
			expected: []string{"openapi-generator", "openapi-generator-cli"},
		},
	}

	// Iterate over test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function
			result := r.LoadExcludedRepos(tt.filename)

			// Compare the result with the expected value
			if !equal(result, tt.expected) {
				t.Errorf("LoadExcludedRepos(%s) = %v; want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

// Helper function to compare two slices.
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
