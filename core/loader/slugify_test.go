package loader

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "Hello World", "hello-world"},
		{"underscores", "pre_build_step", "pre-build-step"},
		{"special chars", "Build & Test!", "build-test"},
		{"consecutive hyphens", "build---test", "build-test"},
		{"leading/trailing hyphens", "-build-test-", "build-test"},
		{"already slugified", "secure-build", "secure-build"},
		{"mixed case with numbers", "Stage 2 Build", "stage-2-build"},
		{"empty string", "", ""},
		{"only special chars", "!@#$%", ""},
		{"spaces and tabs", "  hello   world  ", "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			if result != tt.expected {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
