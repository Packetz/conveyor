package loader

// YAMLPipeline is the top-level YAML pipeline representation.
type YAMLPipeline struct {
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description"`
	Version       string            `yaml:"version"`
	Triggers      []YAMLTrigger     `yaml:"triggers"`
	Environment   *YAMLEnvironment  `yaml:"environment"`
	Cache         *YAMLCache        `yaml:"cache"`
	Stages        []YAMLStage       `yaml:"stages"`
	Notifications interface{}       `yaml:"notifications"`
	Artifacts     interface{}       `yaml:"artifacts"`
}

// YAMLEnvironment holds environment variable configuration.
type YAMLEnvironment struct {
	Variables map[string]string `yaml:"variables"`
}

// YAMLTrigger represents a pipeline trigger.
type YAMLTrigger struct {
	Type     string   `yaml:"type"`
	Branches []string `yaml:"branches"`
	Events   []string `yaml:"events"`
	Paths    []string `yaml:"paths"`
}

// YAMLCache represents cache configuration.
type YAMLCache struct {
	Key    string   `yaml:"key"`
	Paths  []string `yaml:"paths"`
	Policy string   `yaml:"policy"`
}

// YAMLStage represents a pipeline stage.
type YAMLStage struct {
	Name  string     `yaml:"name"`
	Needs []string   `yaml:"needs"`
	When  *YAMLWhen  `yaml:"when"`
	Steps []YAMLStep `yaml:"steps"`
}

// YAMLStep represents a step within a stage.
type YAMLStep struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Type        string                 `yaml:"type"`
	Run         string                 `yaml:"run"`
	Plugin      string                 `yaml:"plugin"`
	Image       string                 `yaml:"image"`
	Environment map[string]string      `yaml:"environment"`
	Config      map[string]interface{} `yaml:"config"`
	When        *YAMLWhen              `yaml:"when"`
	Retry       *YAMLRetry             `yaml:"retry"`
	Timeout     string                 `yaml:"timeout"`
	Cache       *YAMLCache             `yaml:"cache"`
	DependsOn   []string               `yaml:"depends_on"`
	Outputs     map[string]string      `yaml:"outputs"`
}

// YAMLWhen represents conditional execution configuration.
type YAMLWhen struct {
	Branch  string `yaml:"branch"`
	Status  string `yaml:"status"`
	Custom  string `yaml:"custom"`
	Pattern string `yaml:"pattern"`
}

// YAMLRetry represents retry configuration.
type YAMLRetry struct {
	MaxAttempts        int    `yaml:"max_attempts"`
	Interval           string `yaml:"interval"`
	ExponentialBackoff bool   `yaml:"exponential_backoff"`
}
