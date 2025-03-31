package security

import (
	"context"
	"fmt"
	"time"

	"github.com/chip/conveyor/core"
)

// SecurityPlugin implements the Plugin interface for security scanning
type SecurityPlugin struct {
	config SecurityConfig
}

// SecurityConfig represents the security plugin configuration
type SecurityConfig struct {
	VulnerabilityScan VulnerabilityConfig `json:"vulnerabilityScan"`
	SecretScan        SecretConfig        `json:"secretScan"`
	LicenseScan       LicenseConfig       `json:"licenseScan"`
}

// VulnerabilityConfig represents the vulnerability scan configuration
type VulnerabilityConfig struct {
	Enabled     bool     `json:"enabled"`
	Threshold   string   `json:"threshold"`
	ExcludeDeps []string `json:"excludeDeps"`
}

// SecretConfig represents the secret scan configuration
type SecretConfig struct {
	Enabled  bool     `json:"enabled"`
	Patterns []string `json:"patterns"`
}

// LicenseConfig represents the license scan configuration
type LicenseConfig struct {
	Enabled     bool     `json:"enabled"`
	AllowedList []string `json:"allowedList"`
	BlockedList []string `json:"blockedList"`
}

// Scan represents a security scan
type Scan struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	PipelineID    string                 `json:"pipelineId"`
	JobID         string                 `json:"jobId"`
	Status        string                 `json:"status"`
	Timestamp     time.Time              `json:"timestamp"`
	FindingsCount int                    `json:"findingsCount"`
	HighCount     int                    `json:"highCount,omitempty"`
	MediumCount   int                    `json:"mediumCount,omitempty"`
	LowCount      int                    `json:"lowCount,omitempty"`
	Findings      []Finding              `json:"findings,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Finding represents a security finding
type Finding struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Package     string                 `json:"package,omitempty"`
	Version     string                 `json:"version,omitempty"`
	FixVersion  string                 `json:"fixVersion,omitempty"`
	Path        string                 `json:"path,omitempty"`
	LineNumber  int                    `json:"lineNumber,omitempty"`
	Context     string                 `json:"context,omitempty"`
	License     string                 `json:"license,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewSecurityPlugin creates a new security plugin
func NewSecurityPlugin() *SecurityPlugin {
	return &SecurityPlugin{
		config: SecurityConfig{
			VulnerabilityScan: VulnerabilityConfig{
				Enabled:     true,
				Threshold:   "medium",
				ExcludeDeps: []string{"dev-dependencies"},
			},
			SecretScan: SecretConfig{
				Enabled:  true,
				Patterns: []string{"api_key", "password", "secret", "key"},
			},
			LicenseScan: LicenseConfig{
				Enabled:     true,
				AllowedList: []string{"MIT", "Apache-2.0", "BSD-3-Clause"},
				BlockedList: []string{"GPL-3.0"},
			},
		},
	}
}

// GetManifest returns the plugin manifest
func (p *SecurityPlugin) GetManifest() core.PluginManifest {
	return core.PluginManifest{
		Name:        "security",
		Version:     "1.0.0",
		Description: "Security scanning plugin for vulnerability, secret, and license scanning",
		Author:      "Conveyor Team",
		Type:        "scanner",
		StepTypes:   []string{"vulnerability-scan", "secret-scan", "license-scan"},
	}
}

// Execute runs a security scan
func (p *SecurityPlugin) Execute(ctx context.Context, step core.Step) (map[string]interface{}, error) {
	scanID := fmt.Sprintf("scan-%d", time.Now().Unix())
	
	switch step.Type {
	case "vulnerability-scan":
		return p.executeVulnerabilityScan(ctx, scanID, step)
	case "secret-scan":
		return p.executeSecretScan(ctx, scanID, step)
	case "license-scan":
		return p.executeLicenseScan(ctx, scanID, step)
	default:
		return nil, fmt.Errorf("unsupported step type: %s", step.Type)
	}
}

// executeVulnerabilityScan runs a vulnerability scan
func (p *SecurityPlugin) executeVulnerabilityScan(ctx context.Context, scanID string, step core.Step) (map[string]interface{}, error) {
	if !p.config.VulnerabilityScan.Enabled {
		return map[string]interface{}{
			"status": "skipped",
			"reason": "vulnerability scan is disabled",
		}, nil
	}
	
	// Simulate scanning for vulnerabilities
	time.Sleep(1 * time.Second)
	
	// Sample findings for demonstration
	findings := []Finding{
		{
			ID:          "CVE-2021-1234",
			Type:        "vulnerability",
			Title:       "Prototype Pollution",
			Description: "Prototype pollution vulnerability in lodash",
			Severity:    "high",
			Package:     "lodash",
			Version:     "4.17.20",
			FixVersion:  "4.17.21",
		},
		{
			ID:          "CVE-2021-5678",
			Type:        "vulnerability",
			Title:       "Memory Leak",
			Description: "Memory leak in Express.js",
			Severity:    "high",
			Package:     "express",
			Version:     "4.17.1",
			FixVersion:  "4.17.2",
		},
		{
			ID:          "CVE-2022-1111",
			Type:        "vulnerability",
			Title:       "SQL Injection",
			Description: "SQL injection vulnerability in sequelize",
			Severity:    "medium",
			Package:     "sequelize",
			Version:     "6.6.5",
			FixVersion:  "6.6.6",
		},
		{
			ID:          "CVE-2022-2222",
			Type:        "vulnerability",
			Title:       "XSS Vulnerability",
			Description: "Cross-site scripting vulnerability in react",
			Severity:    "medium",
			Package:     "react",
			Version:     "17.0.1",
			FixVersion:  "17.0.2",
		},
	}
	
	scan := Scan{
		ID:            scanID,
		Type:          "vulnerability",
		PipelineID:    step.Config["pipelineId"].(string),
		JobID:         step.Config["jobId"].(string),
		Status:        "completed",
		Timestamp:     time.Now(),
		FindingsCount: len(findings),
		HighCount:     2,
		MediumCount:   2,
		LowCount:      0,
		Findings:      findings,
	}
	
	return map[string]interface{}{
		"scan": scan,
	}, nil
}

// executeSecretScan runs a secret scan
func (p *SecurityPlugin) executeSecretScan(ctx context.Context, scanID string, step core.Step) (map[string]interface{}, error) {
	if !p.config.SecretScan.Enabled {
		return map[string]interface{}{
			"status": "skipped",
			"reason": "secret scan is disabled",
		}, nil
	}
	
	// Simulate scanning for secrets
	time.Sleep(1 * time.Second)
	
	// Sample findings for demonstration
	findings := []Finding{
		{
			ID:          "SECRET-1",
			Type:        "secret",
			Title:       "API Key Exposure",
			Description: "API key exposed in code",
			Severity:    "high",
			Path:        "src/config.js",
			LineNumber:  42,
			Context:     "const apiKey = 'abcdef123456';",
		},
	}
	
	scan := Scan{
		ID:            scanID,
		Type:          "secret",
		PipelineID:    step.Config["pipelineId"].(string),
		JobID:         step.Config["jobId"].(string),
		Status:        "completed",
		Timestamp:     time.Now(),
		FindingsCount: len(findings),
		Findings:      findings,
	}
	
	return map[string]interface{}{
		"scan": scan,
	}, nil
}

// executeLicenseScan runs a license scan
func (p *SecurityPlugin) executeLicenseScan(ctx context.Context, scanID string, step core.Step) (map[string]interface{}, error) {
	if !p.config.LicenseScan.Enabled {
		return map[string]interface{}{
			"status": "skipped",
			"reason": "license scan is disabled",
		}, nil
	}
	
	// Simulate scanning for licenses
	time.Sleep(1 * time.Second)
	
	// Sample findings for demonstration
	findings := []Finding{
		{
			ID:          "LICENSE-1",
			Type:        "license",
			Title:       "GPL-3.0 License",
			Description: "GPL-3.0 license not allowed",
			Severity:    "high",
			Package:     "some-gpl-package",
			Version:     "1.0.0",
			License:     "GPL-3.0",
		},
		{
			ID:          "LICENSE-2",
			Type:        "license",
			Title:       "Unknown License",
			Description: "Package has an unknown license",
			Severity:    "medium",
			Package:     "unknown-license-package",
			Version:     "0.1.0",
			License:     "UNKNOWN",
		},
	}
	
	scan := Scan{
		ID:            scanID,
		Type:          "license",
		PipelineID:    step.Config["pipelineId"].(string),
		JobID:         step.Config["jobId"].(string),
		Status:        "completed",
		Timestamp:     time.Now(),
		FindingsCount: len(findings),
		HighCount:     1,
		MediumCount:   1,
		LowCount:      0,
		Findings:      findings,
	}
	
	return map[string]interface{}{
		"scan": scan,
	}, nil
}

// GetConfig returns the current plugin configuration
func (p *SecurityPlugin) GetConfig() SecurityConfig {
	return p.config
}

// UpdateConfig updates the plugin configuration
func (p *SecurityPlugin) UpdateConfig(config SecurityConfig) {
	p.config = config
} 