package loader

import (
	"os"
	"testing"
)

func TestParse_ValidMinimal(t *testing.T) {
	data, err := os.ReadFile("testdata/valid/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	p, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if p.Name != "minimal-pipeline" {
		t.Errorf("Name = %q, want %q", p.Name, "minimal-pipeline")
	}
	if len(p.Stages) != 1 {
		t.Fatalf("len(Stages) = %d, want 1", len(p.Stages))
	}
	if p.Stages[0].Name != "build" {
		t.Errorf("Stages[0].Name = %q, want %q", p.Stages[0].Name, "build")
	}
	if len(p.Stages[0].Steps) != 1 {
		t.Fatalf("len(Steps) = %d, want 1", len(p.Stages[0].Steps))
	}
	if p.Stages[0].Steps[0].Run != "go build ./..." {
		t.Errorf("Steps[0].Run = %q, want %q", p.Stages[0].Steps[0].Run, "go build ./...")
	}
}

func TestParse_BadSyntax(t *testing.T) {
	data, err := os.ReadFile("testdata/invalid/bad-syntax.yaml")
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	_, err = Parse(data)
	if err == nil {
		t.Error("Parse() expected error for bad YAML syntax, got nil")
	}
}
