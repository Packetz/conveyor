package loader

import (
	"time"

	"github.com/chip/conveyor/core"
)

// Convert transforms a YAMLPipeline into a core.Pipeline with the given ID.
func Convert(p *YAMLPipeline, id string) (*core.Pipeline, error) {
	nameToSlug := make(map[string]string)
	for _, s := range p.Stages {
		nameToSlug[s.Name] = Slugify(s.Name)
	}

	now := time.Now()
	pipeline := &core.Pipeline{
		ID:          id,
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	for _, t := range p.Triggers {
		pipeline.Triggers = append(pipeline.Triggers, core.Trigger{
			Type:     t.Type,
			Branches: t.Branches,
			Events:   t.Events,
			Paths:    t.Paths,
		})
	}

	if p.Cache != nil {
		pipeline.Cache = &core.CacheConfig{
			Key:    p.Cache.Key,
			Paths:  p.Cache.Paths,
			Policy: p.Cache.Policy,
		}
	}

	if p.Environment != nil {
		pipeline.Environment = p.Environment.Variables
	}

	for _, ys := range p.Stages {
		stageID := Slugify(ys.Name)

		stage := core.Stage{
			ID:   stageID,
			Name: ys.Name,
		}

		for _, need := range ys.Needs {
			if slug, ok := nameToSlug[need]; ok {
				stage.Needs = append(stage.Needs, slug)
			}
		}

		if ys.When != nil {
			stage.When = &core.ConditionalExecution{
				Branch:  ys.When.Branch,
				Status:  ys.When.Status,
				Custom:  ys.When.Custom,
				Pattern: ys.When.Pattern,
			}
		}

		for _, yst := range ys.Steps {
			stepID := Slugify(stageID + "-" + yst.Name)

			step := core.Step{
				ID:          stepID,
				Name:        yst.Name,
				Command:     yst.Run,
				Plugin:      yst.Plugin,
				Image:       yst.Image,
				Environment: yst.Environment,
				Config:      yst.Config,
				Timeout:     yst.Timeout,
				DependsOn:   yst.DependsOn,
				Outputs:     yst.Outputs,
			}

			if yst.Type != "" {
				step.Type = yst.Type
			} else if yst.Plugin != "" {
				step.Type = "plugin"
			} else {
				step.Type = "script"
			}

			if yst.Description != "" {
				if step.Metadata == nil {
					step.Metadata = make(map[string]interface{})
				}
				step.Metadata["description"] = yst.Description
			}

			if yst.When != nil {
				step.When = &core.ConditionalExecution{
					Branch:  yst.When.Branch,
					Status:  yst.When.Status,
					Custom:  yst.When.Custom,
					Pattern: yst.When.Pattern,
				}
			}

			if yst.Retry != nil {
				step.Retry = &core.RetryConfig{
					MaxAttempts:        yst.Retry.MaxAttempts,
					Interval:           yst.Retry.Interval,
					ExponentialBackoff: yst.Retry.ExponentialBackoff,
				}
			}

			if yst.Cache != nil {
				step.Cache = &core.CacheConfig{
					Key:    yst.Cache.Key,
					Paths:  yst.Cache.Paths,
					Policy: yst.Cache.Policy,
				}
			}

			stage.Steps = append(stage.Steps, step)
		}

		pipeline.Stages = append(pipeline.Stages, stage)
	}

	return pipeline, nil
}
