package loader

import "gopkg.in/yaml.v3"

// Parse unmarshals YAML bytes into a YAMLPipeline struct.
func Parse(data []byte) (*YAMLPipeline, error) {
	var p YAMLPipeline
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
