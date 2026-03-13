package loader

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chip/conveyor/core"
)

// PipelineLoader loads pipeline definitions from YAML files.
type PipelineLoader struct {
	engine       *core.PipelineEngine
	pipelinesDir string
}

// LoadResult contains the results of loading pipelines from a directory.
type LoadResult struct {
	Loaded   []*core.Pipeline
	Warnings map[string][]string
	Errors   map[string]error
}

// NewPipelineLoader creates a new PipelineLoader.
func NewPipelineLoader(engine *core.PipelineEngine, pipelinesDir string) *PipelineLoader {
	return &PipelineLoader{
		engine:       engine,
		pipelinesDir: pipelinesDir,
	}
}

// LoadDirectory scans the configured directory and loads all YAML pipeline files.
func (l *PipelineLoader) LoadDirectory() (*LoadResult, error) {
	result := &LoadResult{
		Warnings: make(map[string][]string),
		Errors:   make(map[string]error),
	}

	if _, err := os.Stat(l.pipelinesDir); os.IsNotExist(err) {
		log.Printf("Pipeline directory %q does not exist, skipping", l.pipelinesDir)
		return result, nil
	}

	var files []string
	for _, pattern := range []string{"*.yaml", "*.yml"} {
		matches, err := filepath.Glob(filepath.Join(l.pipelinesDir, pattern))
		if err != nil {
			return nil, fmt.Errorf("failed to glob %s: %w", pattern, err)
		}
		files = append(files, matches...)
	}

	sort.Strings(files)

	for _, path := range files {
		filename := filepath.Base(path)
		pipeline, warnings, err := l.LoadFile(path)
		if err != nil {
			result.Errors[filename] = err
			continue
		}
		if len(warnings) > 0 {
			result.Warnings[filename] = warnings
		}
		result.Loaded = append(result.Loaded, pipeline)
	}

	return result, nil
}

// LoadFile loads a single pipeline from a YAML file.
func (l *PipelineLoader) LoadFile(path string) (*core.Pipeline, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	base := filepath.Base(path)
	ext := filepath.Ext(base)
	id := strings.TrimSuffix(base, ext)

	return l.loadPipeline(data, id)
}

// LoadFromBytes loads a pipeline from raw YAML bytes with a given ID.
func (l *PipelineLoader) LoadFromBytes(data []byte, id string) (*core.Pipeline, []string, error) {
	return l.loadPipeline(data, id)
}

func (l *PipelineLoader) loadPipeline(data []byte, id string) (*core.Pipeline, []string, error) {
	yp, err := Parse(data)
	if err != nil {
		return nil, nil, fmt.Errorf("YAML parse error: %w", err)
	}

	warnings, err := Validate(yp)
	if err != nil {
		return nil, warnings, err
	}

	pipeline, err := Convert(yp, id)
	if err != nil {
		return nil, warnings, fmt.Errorf("conversion error: %w", err)
	}

	if err := l.engine.CreatePipeline(pipeline); err != nil {
		return nil, warnings, fmt.Errorf("failed to register pipeline: %w", err)
	}

	return pipeline, warnings, nil
}
