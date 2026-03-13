package loader

import (
	"fmt"
	"strings"
)

// Validate checks a YAMLPipeline for errors and returns warnings for unsupported fields.
func Validate(p *YAMLPipeline) ([]string, error) {
	var errs []string
	var warnings []string

	if strings.TrimSpace(p.Name) == "" {
		errs = append(errs, "pipeline name is required")
	}

	if len(p.Stages) == 0 {
		errs = append(errs, "pipeline must have at least one stage")
	}

	stageNames := make(map[string]bool)
	slugSeen := make(map[string]string)

	for i, stage := range p.Stages {
		if strings.TrimSpace(stage.Name) == "" {
			errs = append(errs, fmt.Sprintf("stage %d: name is required", i+1))
			continue
		}

		slug := Slugify(stage.Name)
		if prevName, exists := slugSeen[slug]; exists {
			errs = append(errs, fmt.Sprintf("stage %q: duplicate slugified ID %q (conflicts with stage %q)", stage.Name, slug, prevName))
		}
		slugSeen[slug] = stage.Name
		stageNames[stage.Name] = true

		if len(stage.Steps) == 0 {
			errs = append(errs, fmt.Sprintf("stage %q: must have at least one step", stage.Name))
		}

		for j, step := range stage.Steps {
			if strings.TrimSpace(step.Name) == "" {
				errs = append(errs, fmt.Sprintf("stage %q, step %d: name is required", stage.Name, j+1))
				continue
			}

			hasRun := strings.TrimSpace(step.Run) != ""
			hasPlugin := strings.TrimSpace(step.Plugin) != ""
			if !hasRun && !hasPlugin {
				errs = append(errs, fmt.Sprintf("stage %q, step %q: must have either 'run' or 'plugin'", stage.Name, step.Name))
			}
			if hasRun && hasPlugin {
				errs = append(errs, fmt.Sprintf("stage %q, step %q: cannot have both 'run' and 'plugin'", stage.Name, step.Name))
			}
		}
	}

	for _, stage := range p.Stages {
		for _, need := range stage.Needs {
			if !stageNames[need] {
				errs = append(errs, fmt.Sprintf("stage %q: needs references unknown stage %q", stage.Name, need))
			}
		}
	}

	if err := detectCycles(p.Stages); err != nil {
		errs = append(errs, err.Error())
	}

	if strings.TrimSpace(p.Version) != "" {
		warnings = append(warnings, "field 'version' is not yet supported and will be ignored")
	}
	if p.Notifications != nil {
		warnings = append(warnings, "field 'notifications' is not yet supported and will be ignored")
	}
	if p.Artifacts != nil {
		warnings = append(warnings, "field 'artifacts' is not yet supported and will be ignored")
	}

	if len(errs) > 0 {
		return warnings, fmt.Errorf("validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return warnings, nil
}

func detectCycles(stages []YAMLStage) error {
	adj := make(map[string][]string)
	for _, s := range stages {
		adj[s.Name] = s.Needs
	}

	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)

	state := make(map[string]int)

	var visit func(name string, path []string) error
	visit = func(name string, path []string) error {
		if state[name] == visited {
			return nil
		}
		if state[name] == visiting {
			return fmt.Errorf("circular dependency detected: %s -> %s", strings.Join(path, " -> "), name)
		}

		state[name] = visiting
		path = append(path, name)

		for _, dep := range adj[name] {
			if err := visit(dep, path); err != nil {
				return err
			}
		}

		state[name] = visited
		return nil
	}

	for _, s := range stages {
		if state[s.Name] == unvisited {
			if err := visit(s.Name, nil); err != nil {
				return err
			}
		}
	}

	return nil
}
