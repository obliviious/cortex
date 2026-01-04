package config

import (
	"fmt"
	"strings"
)

// ExpandPrompt replaces {{outputs.<task-name>}} placeholders in a prompt
// with actual output values from completed tasks.
//
// Example:
//
//	prompt: "Based on: {{outputs.analyze}}\nImplement changes."
//	outputs: {"analyze": "Found 3 issues..."}
//	result: "Based on: Found 3 issues...\nImplement changes."
func ExpandPrompt(prompt string, outputs map[string]string) string {
	result := prompt

	// Find and replace all {{outputs.X}} patterns
	matches := templateVarRegex.FindAllStringSubmatch(prompt, -1)
	for _, match := range matches {
		placeholder := match[0] // Full match: {{outputs.taskname}}
		taskName := match[1]    // Captured group: taskname

		if output, exists := outputs[taskName]; exists {
			result = strings.Replace(result, placeholder, output, -1)
		}
		// If output doesn't exist, leave placeholder as-is (validation should catch this)
	}

	return result
}

// ExtractTemplateVars returns all task names referenced in {{outputs.X}} patterns.
func ExtractTemplateVars(prompt string) []string {
	matches := templateVarRegex.FindAllStringSubmatch(prompt, -1)
	var tasks []string
	seen := make(map[string]bool)

	for _, match := range matches {
		taskName := match[1]
		if !seen[taskName] {
			tasks = append(tasks, taskName)
			seen[taskName] = true
		}
	}

	return tasks
}

// ValidateTemplateOutputs checks that all required outputs are available.
// Returns an error if any referenced output is missing.
func ValidateTemplateOutputs(prompt string, outputs map[string]string) error {
	required := ExtractTemplateVars(prompt)
	var missing []string

	for _, taskName := range required {
		if _, exists := outputs[taskName]; !exists {
			missing = append(missing, taskName)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing outputs for template variables: %v", missing)
	}
	return nil
}
