package security

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/chip/conveyor/core"
)

// SecurityPlugin provides security scanning functionality
type SecurityPlugin struct {
	config SecurityConfig
}

// SecurityConfig holds configuration for the security plugin
type SecurityConfig struct {
	SeverityThreshold string   `json:"severityThreshold"`
	IgnorePatterns    []string `json:"ignorePatterns"`
	ScanTypes         []string `json:"scanTypes"`
	CustomRules       []Rule   `json:"customRules"`
	FailOnViolation   bool     `json:"failOnViolation"`
	GenerateSBOM      bool     `json:"generateSBOM"`
	SBOMFormat        string   `json:"sbomFormat"`
	OutputDir         string   `json:"outputDir"`
}

// Rule represents a security rule to check
type Rule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Pattern     string `json:"pattern"`
}

// ScanResult holds the result of a security scan
type ScanResult struct {
	Findings    []Finding     `json:"findings"`
	Summary     ScanSummary   `json:"summary"`
	SBOM        *SBOM         `json:"sbom,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
	Environment string        `json:"environment"`
	Duration    time.Duration `json:"duration"`
}

// Finding represents a security issue found during scanning
type Finding struct {
	RuleID      string `json:"ruleId"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Location    string `json:"location"`
	LineNumber  int    `json:"lineNumber"`
	Context     string `json:"context"`
	Remediation string `json:"remediation,omitempty"`
}

// ScanSummary provides an overview of the scan results
type ScanSummary struct {
	TotalFiles     int            `json:"totalFiles"`
	FilesScanned   int            `json:"filesScanned"`
	FilesSkipped   int            `json:"filesSkipped"`
	TotalFindings  int            `json:"totalFindings"`
	FindingsBySeverity map[string]int `json:"findingsBySeverity"`
	PassedCheck    bool           `json:"passedCheck"`
}

// SBOM represents a Software Bill of Materials
type SBOM struct {
	Components []Component `json:"components"`
	Format     string      `json:"format"`
	Version    string      `json:"version"`
}

// Component represents a component in an SBOM
type Component struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Type        string    `json:"type"`
	License     string    `json:"license"`
	Source      string    `json:"source"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`
}

// Vulnerability represents a known vulnerability in a component
type Vulnerability struct {
	ID         string `json:"id"`
	Severity   string `json:"severity"`
	CVSS       float64 `json:"cvss"`
	Description string `json:"description"`
	FixedIn    string `json:"fixedIn,omitempty"`
}

// NewSecurityPlugin creates a new security plugin
func NewSecurityPlugin(configFile string) (*SecurityPlugin, error) {
	// Default configuration
	config := SecurityConfig{
		SeverityThreshold: "HIGH",
		IgnorePatterns:    []string{"node_modules/", "vendor/", ".git/"},
		ScanTypes:         []string{"secret", "vulnerability", "license", "code"},
		FailOnViolation:   true,
		GenerateSBOM:      true,
		SBOMFormat:        "cyclonedx",
		OutputDir:         "security-reports",
	}

	// Load configuration from file if provided
	if configFile != "" {
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Load default rules
	config.CustomRules = append(config.CustomRules, getDefaultRules()...)

	return &SecurityPlugin{
		config: config,
	}, nil
}

// GetManifest returns the plugin manifest
func (p *SecurityPlugin) GetManifest() core.PluginManifest {
	return core.PluginManifest{
		Name:        "security-scanner",
		Version:     "2.0.0",
		Description: "Comprehensive security scanner for code, dependencies, and configurations",
		Author:      "Conveyor Team",
		StepTypes:   []string{"security-scan", "vulnerability-scan", "secret-scan", "sbom-generate"},
	}
}

// Execute runs the security scan
func (p *SecurityPlugin) Execute(ctx context.Context, step core.Step) (map[string]interface{}, error) {
	startTime := time.Now()

	// Customize configuration with step-specific settings
	config := p.config
	if scanTypes, ok := step.Config["scanTypes"].([]interface{}); ok {
		config.ScanTypes = make([]string, len(scanTypes))
		for i, t := range scanTypes {
			config.ScanTypes[i] = t.(string)
		}
	}

	if threshold, ok := step.Config["severityThreshold"].(string); ok {
		config.SeverityThreshold = threshold
	}

	if failOnViolation, ok := step.Config["failOnViolation"].(bool); ok {
		config.FailOnViolation = failOnViolation
	}

	if generateSBOM, ok := step.Config["generateSBOM"].(bool); ok {
		config.GenerateSBOM = generateSBOM
	}

	// Get the target directory to scan
	targetDir := "."
	if dir, ok := step.Config["targetDir"].(string); ok {
		targetDir = dir
	}

	// Run the scan
	result, err := p.scanDirectory(ctx, targetDir, config)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	result.Duration = time.Since(startTime)

	// Generate reports
	if err := p.generateReports(result, config); err != nil {
		return map[string]interface{}{
			"success":  false,
			"result":   result,
			"error":    err.Error(),
			"duration": result.Duration.String(),
		}, fmt.Errorf("failed to generate reports: %w", err)
	}

	// Check if scan should fail based on findings
	if !result.Summary.PassedCheck && config.FailOnViolation {
		return map[string]interface{}{
			"success":  false,
			"result":   result,
			"error":    fmt.Sprintf("security scan failed with %d findings above threshold", countSevereFindings(result, config.SeverityThreshold)),
			"duration": result.Duration.String(),
		}, fmt.Errorf("security violations found")
	}

	return map[string]interface{}{
		"success":  true,
		"result":   result,
		"duration": result.Duration.String(),
	}, nil
}

// scanDirectory performs security scanning on a directory
func (p *SecurityPlugin) scanDirectory(ctx context.Context, targetDir string, config SecurityConfig) (*ScanResult, error) {
	result := &ScanResult{
		Findings:  []Finding{},
		Timestamp: time.Now(),
		Summary: ScanSummary{
			FindingsBySeverity: make(map[string]int),
			PassedCheck:        true,
		},
	}

	// Get list of files to scan
	files, err := p.getFilesToScan(targetDir, config.IgnorePatterns)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	result.Summary.TotalFiles = len(files)

	// Scan for secrets and code vulnerabilities
	if contains(config.ScanTypes, "secret") || contains(config.ScanTypes, "code") {
		codeFindings, scannedFiles, err := p.scanForCodeIssues(files, config.CustomRules)
		if err != nil {
			return nil, fmt.Errorf("code scan failed: %w", err)
		}

		result.Findings = append(result.Findings, codeFindings...)
		result.Summary.FilesScanned += scannedFiles
	}

	// Scan for dependencies and vulnerabilities
	if contains(config.ScanTypes, "vulnerability") {
		vulnFindings, components, err := p.scanForVulnerabilities(targetDir)
		if err != nil {
			return nil, fmt.Errorf("vulnerability scan failed: %w", err)
		}

		result.Findings = append(result.Findings, vulnFindings...)
		
		// Generate SBOM if requested
		if config.GenerateSBOM {
			result.SBOM = &SBOM{
				Components: components,
				Format:     config.SBOMFormat,
				Version:    "1.0",
			}
		}
	}

	// Scan for license compliance
	if contains(config.ScanTypes, "license") {
		licenseFindings, err := p.scanForLicenseIssues(targetDir)
		if err != nil {
			return nil, fmt.Errorf("license scan failed: %w", err)
		}

		result.Findings = append(result.Findings, licenseFindings...)
	}

	// Generate summary statistics
	result.Summary.TotalFindings = len(result.Findings)
	result.Summary.FilesSkipped = result.Summary.TotalFiles - result.Summary.FilesScanned

	// Count findings by severity
	for _, finding := range result.Findings {
		result.Summary.FindingsBySeverity[finding.Severity]++
		
		// Check if this finding exceeds our threshold
		if exceedsSeverityThreshold(finding.Severity, config.SeverityThreshold) {
			result.Summary.PassedCheck = false
		}
	}

	return result, nil
}

// getFilesToScan returns a list of files to scan, excluding ignored patterns
func (p *SecurityPlugin) getFilesToScan(targetDir string, ignorePatterns []string) ([]string, error) {
	var files []string

	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			for _, pattern := range ignorePatterns {
				matched, err := filepath.Match(pattern, info.Name())
				if err != nil {
					return err
				}
				if matched {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check if file should be ignored
		for _, pattern := range ignorePatterns {
			matched, err := filepath.Match(pattern, path)
			if err != nil {
				return err
			}
			if matched {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

// scanForCodeIssues scans files for code issues and secrets
func (p *SecurityPlugin) scanForCodeIssues(files []string, rules []Rule) ([]Finding, int, error) {
	var findings []Finding
	scannedFiles := 0

	for _, file := range files {
		// Skip binary files
		if isBinaryFile(file) {
			continue
		}

		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, scannedFiles, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		content := string(data)
		lines := strings.Split(content, "\n")
		scannedFiles++

		// Check each rule against the file
		for _, rule := range rules {
			regex, err := regexp.Compile(rule.Pattern)
			if err != nil {
				return nil, scannedFiles, fmt.Errorf("invalid rule pattern %s: %w", rule.ID, err)
			}

			for i, line := range lines {
				if regex.MatchString(line) {
					findings = append(findings, Finding{
						RuleID:      rule.ID,
						Severity:    rule.Severity,
						Description: rule.Description,
						Location:    file,
						LineNumber:  i + 1,
						Context:     sanitizeContext(line),
						Remediation: getRemediation(rule.ID),
					})
				}
			}
		}
	}

	return findings, scannedFiles, nil
}

// scanForVulnerabilities scans for vulnerabilities in dependencies
func (p *SecurityPlugin) scanForVulnerabilities(targetDir string) ([]Finding, []Component, error) {
	// This is a simplified implementation - in a real plugin we would
	// use tools like OWASP Dependency Check, Trivy, or Grype
	
	var findings []Finding
	var components []Component

	// Check for package.json (Node.js)
	pkgJsonPath := filepath.Join(targetDir, "package.json")
	if _, err := os.Stat(pkgJsonPath); err == nil {
		npmComponents, npmFindings := p.scanNodeDependencies(pkgJsonPath)
		components = append(components, npmComponents...)
		findings = append(findings, npmFindings...)
	}

	// Check for go.mod (Go)
	goModPath := filepath.Join(targetDir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		goComponents, goFindings := p.scanGoDependencies(goModPath)
		components = append(components, goComponents...)
		findings = append(findings, goFindings...)
	}

	// Check for pom.xml (Java/Maven)
	pomPath := filepath.Join(targetDir, "pom.xml")
	if _, err := os.Stat(pomPath); err == nil {
		mavenComponents, mavenFindings := p.scanMavenDependencies(pomPath)
		components = append(components, mavenComponents...)
		findings = append(findings, mavenFindings...)
	}

	return findings, components, nil
}

// scanNodeDependencies scans Node.js dependencies for vulnerabilities
func (p *SecurityPlugin) scanNodeDependencies(packageJsonPath string) ([]Component, []Finding) {
	// Simplified implementation - would use npm audit or other tools in reality
	
	// Mock result with a sample vulnerability
	components := []Component{
		{
			Name:    "axios",
			Version: "0.21.1",
			Type:    "npm",
			License: "MIT",
			Source:  "https://www.npmjs.com/package/axios",
			Vulnerabilities: []Vulnerability{
				{
					ID:          "CVE-2021-3749",
					Severity:    "HIGH",
					CVSS:        8.1,
					Description: "Axios before 0.21.2 allows server-side request forgery",
					FixedIn:     "0.21.2",
				},
			},
		},
	}

	findings := []Finding{
		{
			RuleID:      "VULN-001",
			Severity:    "HIGH",
			Description: "Axios before 0.21.2 allows server-side request forgery (CVE-2021-3749)",
			Location:    packageJsonPath,
			LineNumber:  0, // Would be populated with actual line number in real implementation
			Context:     "\"axios\": \"^0.21.1\"",
			Remediation: "Update axios to version 0.21.2 or later",
		},
	}

	return components, findings
}

// scanGoDependencies scans Go dependencies for vulnerabilities
func (p *SecurityPlugin) scanGoDependencies(goModPath string) ([]Component, []Finding) {
	// In a real implementation, would use govulncheck or similar tools
	
	// Mock data for demonstration
	components := []Component{
		{
			Name:    "github.com/gin-gonic/gin",
			Version: "v1.7.4",
			Type:    "go",
			License: "MIT",
			Source:  "https://github.com/gin-gonic/gin",
		},
	}

	// Empty findings for this example
	findings := []Finding{}

	return components, findings
}

// scanMavenDependencies scans Maven dependencies for vulnerabilities
func (p *SecurityPlugin) scanMavenDependencies(pomPath string) ([]Component, []Finding) {
	// In a real implementation, would use dependency-check or similar tools
	
	// Mock data for demonstration
	components := []Component{
		{
			Name:    "org.springframework:spring-core",
			Version: "5.3.9",
			Type:    "maven",
			License: "Apache-2.0",
			Source:  "https://mvnrepository.com/artifact/org.springframework/spring-core",
		},
	}

	// Empty findings for this example
	findings := []Finding{}

	return components, findings
}

// scanForLicenseIssues scans for license compliance issues
func (p *SecurityPlugin) scanForLicenseIssues(targetDir string) ([]Finding, error) {
	// In a real implementation, would use license scanners
	
	// Mock data for demonstration
	findings := []Finding{
		{
			RuleID:      "LICENSE-001",
			Severity:    "MEDIUM",
			Description: "Detected GPL-3.0 license which may conflict with project requirements",
			Location:    filepath.Join(targetDir, "some-gpl-library/LICENSE"),
			LineNumber:  0,
			Context:     "GNU General Public License v3.0",
			Remediation: "Review license compatibility with legal team",
		},
	}

	return findings, nil
}

// generateReports creates security reports from scan results
func (p *SecurityPlugin) generateReports(result *ScanResult, config SecurityConfig) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate JSON report
	jsonReport, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	jsonReportPath := filepath.Join(config.OutputDir, fmt.Sprintf("security-report-%s.json", time.Now().Format("20060102-150405")))
	if err := ioutil.WriteFile(jsonReportPath, jsonReport, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	// Generate SBOM if requested
	if config.GenerateSBOM && result.SBOM != nil {
		sbomData, err := json.MarshalIndent(result.SBOM, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal SBOM: %w", err)
		}

		sbomPath := filepath.Join(config.OutputDir, fmt.Sprintf("sbom-%s.json", time.Now().Format("20060102-150405")))
		if err := ioutil.WriteFile(sbomPath, sbomData, 0644); err != nil {
			return fmt.Errorf("failed to write SBOM: %w", err)
		}
	}

	return nil
}

// Helper functions

// getDefaultRules returns a list of default security rules
func getDefaultRules() []Rule {
	return []Rule{
		{
			ID:          "SECRET-001",
			Name:        "AWS Access Key",
			Description: "AWS Access Key ID detected",
			Severity:    "CRITICAL",
			Pattern:     "AKIA[0-9A-Z]{16}",
		},
		{
			ID:          "SECRET-002",
			Name:        "AWS Secret Key",
			Description: "AWS Secret Access Key detected",
			Severity:    "CRITICAL",
			Pattern:     "[0-9a-zA-Z/+]{40}",
		},
		{
			ID:          "SECRET-003",
			Name:        "GitHub Token",
			Description: "GitHub Token detected",
			Severity:    "CRITICAL",
			Pattern:     "ghp_[0-9a-zA-Z]{36}",
		},
		{
			ID:          "CODE-001",
			Name:        "Insecure Random",
			Description: "Use of insecure random number generator",
			Severity:    "HIGH",
			Pattern:     "Math\\.random\\(\\)",
		},
		{
			ID:          "CODE-002",
			Name:        "SQL Injection",
			Description: "Potential SQL injection vulnerability",
			Severity:    "HIGH",
			Pattern:     "SELECT.*FROM.*WHERE.*\\$|exec\\(.*\\$",
		},
		{
			ID:          "CODE-003",
			Name:        "Hardcoded IP",
			Description: "Hardcoded IP address",
			Severity:    "MEDIUM",
			Pattern:     "\\b\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\b",
		},
	}
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// isBinaryFile determines if a file is binary
func isBinaryFile(filename string) bool {
	// Simple check based on file extension - in a real plugin this would be more sophisticated
	ext := strings.ToLower(filepath.Ext(filename))
	binaryExts := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".zip", ".tar", ".gz", ".exe", ".dll", ".so"}
	
	for _, binaryExt := range binaryExts {
		if ext == binaryExt {
			return true
		}
	}
	
	return false
}

// sanitizeContext removes sensitive information from context
func sanitizeContext(line string) string {
	// Redact potential secrets - in a real plugin this would be more sophisticated
	redactionPatterns := []string{
		"AKIA[0-9A-Z]{16}",
		"[0-9a-zA-Z/+]{40}",
		"ghp_[0-9a-zA-Z]{36}",
		"-----BEGIN .*?PRIVATE KEY-----",
	}
	
	result := line
	for _, pattern := range redactionPatterns {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		result = regex.ReplaceAllString(result, "[REDACTED]")
	}
	
	return result
}

// getRemediation returns remediation guidance for a rule
func getRemediation(ruleID string) string {
	remediations := map[string]string{
		"SECRET-001": "Remove AWS Access Key from code and use environment variables or AWS IAM roles",
		"SECRET-002": "Remove AWS Secret Key from code and use environment variables or AWS IAM roles",
		"SECRET-003": "Remove GitHub Token from code and use environment variables or GitHub Actions secrets",
		"CODE-001": "Use a cryptographically secure random number generator like crypto.randomBytes() in Node.js",
		"CODE-002": "Use parameterized queries or prepared statements for database operations",
		"CODE-003": "Move IP addresses to configuration files or environment variables",
		"VULN-001": "Update vulnerable dependency to the latest secure version",
		"LICENSE-001": "Review license compatibility with legal team and replace component if necessary",
	}
	
	if remediation, ok := remediations[ruleID]; ok {
		return remediation
	}
	
	return "Review and fix the identified issue"
}

// exceedsSeverityThreshold checks if a severity level exceeds the threshold
func exceedsSeverityThreshold(severity, threshold string) bool {
	severityLevels := map[string]int{
		"CRITICAL": 4,
		"HIGH":     3,
		"MEDIUM":   2,
		"LOW":      1,
		"INFO":     0,
	}
	
	severityValue, ok := severityLevels[severity]
	if !ok {
		return false
	}
	
	thresholdValue, ok := severityLevels[threshold]
	if !ok {
		return false
	}
	
	return severityValue >= thresholdValue
}

// countSevereFindings counts findings that exceed the threshold
func countSevereFindings(result *ScanResult, threshold string) int {
	count := 0
	for _, finding := range result.Findings {
		if exceedsSeverityThreshold(finding.Severity, threshold) {
			count++
		}
	}
	return count
} 