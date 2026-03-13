package loader

import (
	"os"
	"strings"
	"testing"
)

func mustParse(t *testing.T, path string) *YAMLPipeline {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	p, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(%s) error = %v", path, err)
	}
	return p
}

func TestValidate_ValidMinimal(t *testing.T) {
	p := mustParse(t, "testdata/valid/minimal.yaml")
	warnings, err := Validate(p)
	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
	if len(warnings) != 0 {
		t.Errorf("Validate() warnings = %v, want none", warnings)
	}
}

func TestValidate_MissingName(t *testing.T) {
	p := mustParse(t, "testdata/invalid/missing-name.yaml")
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error = %q, want it to mention 'name'", err.Error())
	}
}

func TestValidate_EmptyStages(t *testing.T) {
	p := mustParse(t, "testdata/invalid/empty-stages.yaml")
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "stage") {
		t.Errorf("error = %q, want it to mention 'stage'", err.Error())
	}
}

func TestValidate_NoRunOrPlugin(t *testing.T) {
	p := mustParse(t, "testdata/invalid/no-run-or-plugin.yaml")
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "run") || !strings.Contains(err.Error(), "plugin") {
		t.Errorf("error = %q, want it to mention 'run' and 'plugin'", err.Error())
	}
}

func TestValidate_BothRunAndPlugin(t *testing.T) {
	p := &YAMLPipeline{
		Name: "both",
		Stages: []YAMLStage{
			{
				Name: "build",
				Steps: []YAMLStep{
					{Name: "step", Run: "echo hi", Plugin: "security"},
				},
			},
		},
	}
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
}

func TestValidate_CircularDeps(t *testing.T) {
	p := mustParse(t, "testdata/invalid/circular-deps.yaml")
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("error = %q, want it to mention 'circular'", err.Error())
	}
}

func TestValidate_InvalidNeedsRef(t *testing.T) {
	p := &YAMLPipeline{
		Name: "bad-ref",
		Stages: []YAMLStage{
			{
				Name:  "build",
				Needs: []string{"nonexistent"},
				Steps: []YAMLStep{
					{Name: "step", Run: "echo hi"},
				},
			},
		},
	}
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error = %q, want it to mention 'nonexistent'", err.Error())
	}
}

func TestValidate_DiamondDeps(t *testing.T) {
	p := &YAMLPipeline{
		Name: "diamond",
		Stages: []YAMLStage{
			{Name: "a", Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
			{Name: "b", Needs: []string{"a"}, Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
			{Name: "c", Needs: []string{"a"}, Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
			{Name: "d", Needs: []string{"b", "c"}, Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
		},
	}
	_, err := Validate(p)
	if err != nil {
		t.Errorf("Validate() error = %v, want nil for diamond deps", err)
	}
}

func TestValidate_UnsupportedFieldsWarnings(t *testing.T) {
	p := &YAMLPipeline{
		Name:          "with-extras",
		Version:       "1.0.0",
		Notifications: map[string]interface{}{"type": "slack"},
		Artifacts:     []interface{}{"report.html"},
		Stages: []YAMLStage{
			{Name: "build", Steps: []YAMLStep{{Name: "step", Run: "echo"}}},
		},
	}
	warnings, err := Validate(p)
	if err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
	if len(warnings) != 3 {
		t.Errorf("len(warnings) = %d, want 3", len(warnings))
	}
}

func TestValidate_MissingStageName(t *testing.T) {
	p := &YAMLPipeline{
		Name: "missing-stage-name",
		Stages: []YAMLStage{
			{Steps: []YAMLStep{{Name: "step", Run: "echo"}}},
		},
	}
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
}

func TestValidate_MissingStepName(t *testing.T) {
	p := &YAMLPipeline{
		Name: "missing-step-name",
		Stages: []YAMLStage{
			{Name: "build", Steps: []YAMLStep{{Run: "echo"}}},
		},
	}
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
}

func TestValidate_DuplicateStageSlug(t *testing.T) {
	p := &YAMLPipeline{
		Name: "dup-slugs",
		Stages: []YAMLStage{
			{Name: "Build Test", Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
			{Name: "build-test", Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
		},
	}
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error for duplicate slugified stage IDs, got nil")
	}
}

func TestValidate_EmptyStage(t *testing.T) {
	p := &YAMLPipeline{
		Name: "empty-stage",
		Stages: []YAMLStage{
			{Name: "build", Steps: []YAMLStep{}},
		},
	}
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error for stage with zero steps, got nil")
	}
	if !strings.Contains(err.Error(), "at least one step") {
		t.Errorf("error = %q, want it to mention 'at least one step'", err.Error())
	}
}

func TestValidate_ThreeNodeCycle(t *testing.T) {
	p := &YAMLPipeline{
		Name: "three-node-cycle",
		Stages: []YAMLStage{
			{Name: "a", Needs: []string{"c"}, Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
			{Name: "b", Needs: []string{"a"}, Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
			{Name: "c", Needs: []string{"b"}, Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
		},
	}
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error for 3-node circular dep, got nil")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("error = %q, want it to mention 'circular'", err.Error())
	}
}
