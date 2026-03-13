package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chip/conveyor/core"
)

func newTestEngine() *core.PipelineEngine {
	return core.NewPipelineEngine()
}

func TestLoadFile_ValidMinimal(t *testing.T) {
	engine := newTestEngine()
	l := NewPipelineLoader(engine, "testdata/valid")

	absPath, _ := filepath.Abs("testdata/valid/minimal.yaml")
	pipeline, warnings, err := l.LoadFile(absPath)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if pipeline.ID != "minimal" {
		t.Errorf("ID = %q, want %q", pipeline.ID, "minimal")
	}

	got, err := engine.GetPipeline("minimal")
	if err != nil {
		t.Fatalf("engine.GetPipeline() error = %v", err)
	}
	if got.Name != "minimal-pipeline" {
		t.Errorf("engine pipeline Name = %q, want %q", got.Name, "minimal-pipeline")
	}
}

func TestLoadFile_SecureBuild(t *testing.T) {
	engine := newTestEngine()
	l := NewPipelineLoader(engine, "testdata/valid")

	absPath, _ := filepath.Abs("testdata/valid/secure-build.yaml")
	pipeline, warnings, err := l.LoadFile(absPath)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if len(warnings) != 3 {
		t.Errorf("len(warnings) = %d, want 3, got %v", len(warnings), warnings)
	}
	if len(pipeline.Stages) != 6 {
		t.Errorf("len(Stages) = %d, want 6", len(pipeline.Stages))
	}
}

func TestLoadFile_InvalidYAML(t *testing.T) {
	engine := newTestEngine()
	l := NewPipelineLoader(engine, "testdata/invalid")

	absPath, _ := filepath.Abs("testdata/invalid/bad-syntax.yaml")
	_, _, err := l.LoadFile(absPath)
	if err == nil {
		t.Fatal("LoadFile() expected error for bad YAML, got nil")
	}
}

func TestLoadFromBytes_UsesProvidedID(t *testing.T) {
	engine := newTestEngine()
	l := NewPipelineLoader(engine, "")

	yamlData := []byte(`
name: Some Pipeline
stages:
  - name: build
    steps:
      - name: compile
        run: go build ./...
`)

	pipeline, _, err := l.LoadFromBytes(yamlData, "custom-id")
	if err != nil {
		t.Fatalf("LoadFromBytes() error = %v", err)
	}
	if pipeline.ID != "custom-id" {
		t.Errorf("ID = %q, want %q", pipeline.ID, "custom-id")
	}
}

func TestLoadDirectory_MixedFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loader-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	validYAML := []byte("name: valid\nstages:\n  - name: build\n    steps:\n      - name: step\n        run: echo hi\n")
	os.WriteFile(filepath.Join(tmpDir, "valid.yaml"), validYAML, 0644)

	invalidYAML := []byte("name: invalid\nstages: []\n")
	os.WriteFile(filepath.Join(tmpDir, "invalid.yaml"), invalidYAML, 0644)

	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not yaml"), 0644)

	engine := newTestEngine()
	l := NewPipelineLoader(engine, tmpDir)

	result, err := l.LoadDirectory()
	if err != nil {
		t.Fatalf("LoadDirectory() error = %v", err)
	}

	if len(result.Loaded) != 1 {
		t.Errorf("len(Loaded) = %d, want 1", len(result.Loaded))
	}
	if len(result.Errors) != 1 {
		t.Errorf("len(Errors) = %d, want 1", len(result.Errors))
	}
	if _, ok := result.Errors["invalid.yaml"]; !ok {
		t.Errorf("expected error for invalid.yaml, got errors for: %v", result.Errors)
	}
}

func TestLoadDirectory_Nonexistent(t *testing.T) {
	engine := newTestEngine()
	l := NewPipelineLoader(engine, "/nonexistent/path")

	result, err := l.LoadDirectory()
	if err != nil {
		t.Fatalf("LoadDirectory() error = %v", err)
	}
	if len(result.Loaded) != 0 {
		t.Errorf("len(Loaded) = %d, want 0", len(result.Loaded))
	}
}

func TestLoadFromBytes_DuplicatePipelineID(t *testing.T) {
	engine := newTestEngine()
	l := NewPipelineLoader(engine, "")

	yamlData := []byte("name: Dupe\nstages:\n  - name: build\n    steps:\n      - name: step\n        run: echo\n")

	_, _, err := l.LoadFromBytes(yamlData, "dupe-id")
	if err != nil {
		t.Fatalf("first LoadFromBytes() error = %v", err)
	}

	_, _, err = l.LoadFromBytes(yamlData, "dupe-id")
	if err == nil {
		t.Fatal("second LoadFromBytes() expected error for duplicate pipeline ID, got nil")
	}
}

func TestLoadDirectory_DuplicateExtensions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loader-ext-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	yamlData := []byte("name: Same Pipeline\nstages:\n  - name: build\n    steps:\n      - name: step\n        run: echo\n")
	os.WriteFile(filepath.Join(tmpDir, "build.yaml"), yamlData, 0644)
	os.WriteFile(filepath.Join(tmpDir, "build.yml"), yamlData, 0644)

	engine := newTestEngine()
	l := NewPipelineLoader(engine, tmpDir)

	result, err := l.LoadDirectory()
	if err != nil {
		t.Fatalf("LoadDirectory() error = %v", err)
	}

	if len(result.Loaded) != 1 {
		t.Errorf("len(Loaded) = %d, want 1", len(result.Loaded))
	}
	if len(result.Errors) != 1 {
		t.Errorf("len(Errors) = %d, want 1", len(result.Errors))
	}
}

func TestLoadFromBytes_NameFromYAML(t *testing.T) {
	engine := newTestEngine()
	l := NewPipelineLoader(engine, "")

	yamlData := []byte("name: My Actual Name\nstages:\n  - name: build\n    steps:\n      - name: step\n        run: echo\n")

	pipeline, _, err := l.LoadFromBytes(yamlData, "custom-id")
	if err != nil {
		t.Fatalf("LoadFromBytes() error = %v", err)
	}
	if pipeline.ID != "custom-id" {
		t.Errorf("ID = %q, want %q", pipeline.ID, "custom-id")
	}
	if pipeline.Name != "My Actual Name" {
		t.Errorf("Name = %q, want %q", pipeline.Name, "My Actual Name")
	}
}
