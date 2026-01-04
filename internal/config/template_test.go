package config

import (
	"reflect"
	"strings"
	"testing"
)

// TestExpandPrompt tests prompt template expansion with task outputs.
func TestExpandPrompt(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		outputs map[string]string
		want    string
	}{
		{
			name:    "no template variables",
			prompt:  "Simple prompt with no variables",
			outputs: map[string]string{},
			want:    "Simple prompt with no variables",
		},
		{
			name:   "single template variable",
			prompt: "Based on: {{outputs.task1}}",
			outputs: map[string]string{
				"task1": "result from task1",
			},
			want: "Based on: result from task1",
		},
		{
			name:   "multiple template variables",
			prompt: "Combine {{outputs.task1}} and {{outputs.task2}}",
			outputs: map[string]string{
				"task1": "first result",
				"task2": "second result",
			},
			want: "Combine first result and second result",
		},
		{
			name:   "repeated template variable",
			prompt: "Start: {{outputs.task1}}\nEnd: {{outputs.task1}}",
			outputs: map[string]string{
				"task1": "repeated value",
			},
			want: "Start: repeated value\nEnd: repeated value",
		},
		{
			name:   "template variable with hyphens and underscores",
			prompt: "Use {{outputs.task-1_test}}",
			outputs: map[string]string{
				"task-1_test": "complex name result",
			},
			want: "Use complex name result",
		},
		{
			name:    "missing output - leaves placeholder unchanged",
			prompt:  "Use {{outputs.missing}}",
			outputs: map[string]string{},
			want:    "Use {{outputs.missing}}",
		},
		{
			name:   "multiline output",
			prompt: "Result:\n{{outputs.task1}}\nDone.",
			outputs: map[string]string{
				"task1": "Line 1\nLine 2\nLine 3",
			},
			want: "Result:\nLine 1\nLine 2\nLine 3\nDone.",
		},
		{
			name:   "special characters in output",
			prompt: "Data: {{outputs.task1}}",
			outputs: map[string]string{
				"task1": "Result with $pecial ch@rs & symbols!",
			},
			want: "Data: Result with $pecial ch@rs & symbols!",
		},
		{
			name:   "empty output value",
			prompt: "Value: {{outputs.task1}}",
			outputs: map[string]string{
				"task1": "",
			},
			want: "Value: ",
		},
		{
			name:   "mixed present and missing outputs",
			prompt: "{{outputs.task1}} and {{outputs.missing}}",
			outputs: map[string]string{
				"task1": "present",
			},
			want: "present and {{outputs.missing}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandPrompt(tt.prompt, tt.outputs)
			if got != tt.want {
				t.Errorf("ExpandPrompt() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestExtractTemplateVars tests extraction of template variable names.
func TestExtractTemplateVars(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		want   []string
	}{
		{
			name:   "no template variables",
			prompt: "Plain text prompt",
			want:   []string{},
		},
		{
			name:   "single template variable",
			prompt: "Use {{outputs.task1}}",
			want:   []string{"task1"},
		},
		{
			name:   "multiple different variables",
			prompt: "{{outputs.task1}} and {{outputs.task2}}",
			want:   []string{"task1", "task2"},
		},
		{
			name:   "duplicate variables - returns unique only",
			prompt: "{{outputs.task1}} repeated {{outputs.task1}}",
			want:   []string{"task1"},
		},
		{
			name:   "variables with hyphens and underscores",
			prompt: "{{outputs.task-1}} and {{outputs.task_2}}",
			want:   []string{"task-1", "task_2"},
		},
		{
			name:   "multiple duplicates",
			prompt: "{{outputs.a}} {{outputs.b}} {{outputs.a}} {{outputs.c}} {{outputs.b}}",
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "variable in multiline prompt",
			prompt: "Start\n{{outputs.task1}}\nMiddle\n{{outputs.task2}}\nEnd",
			want:   []string{"task1", "task2"},
		},
		{
			name:   "alphanumeric variable names",
			prompt: "{{outputs.task123}} and {{outputs.task456abc}}",
			want:   []string{"task123", "task456abc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractTemplateVars(tt.prompt)

			// Handle empty slice comparison
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractTemplateVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestValidateTemplateOutputs tests validation that required outputs are available.
func TestValidateTemplateOutputs(t *testing.T) {
	tests := []struct {
		name        string
		prompt      string
		outputs     map[string]string
		wantErr     bool
		wantErrContains string
	}{
		{
			name:    "no template variables - valid",
			prompt:  "Plain prompt",
			outputs: map[string]string{},
			wantErr: false,
		},
		{
			name:   "all required outputs present",
			prompt: "Use {{outputs.task1}}",
			outputs: map[string]string{
				"task1": "result",
			},
			wantErr: false,
		},
		{
			name:   "multiple required outputs all present",
			prompt: "{{outputs.task1}} and {{outputs.task2}}",
			outputs: map[string]string{
				"task1": "result1",
				"task2": "result2",
			},
			wantErr: false,
		},
		{
			name:    "missing single required output",
			prompt:  "Use {{outputs.task1}}",
			outputs: map[string]string{},
			wantErr: true,
			wantErrContains: "missing outputs for template variables",
		},
		{
			name:   "missing one of multiple outputs",
			prompt: "{{outputs.task1}} and {{outputs.task2}}",
			outputs: map[string]string{
				"task1": "result1",
			},
			wantErr: true,
			wantErrContains: "task2",
		},
		{
			name:    "missing all outputs",
			prompt:  "{{outputs.task1}} and {{outputs.task2}}",
			outputs: map[string]string{},
			wantErr: true,
			wantErrContains: "missing outputs",
		},
		{
			name:   "extra outputs present - valid",
			prompt: "Use {{outputs.task1}}",
			outputs: map[string]string{
				"task1": "result1",
				"task2": "extra result",
				"task3": "another extra",
			},
			wantErr: false,
		},
		{
			name:   "duplicate template vars - only checked once",
			prompt: "{{outputs.task1}} repeated {{outputs.task1}}",
			outputs: map[string]string{
				"task1": "result",
			},
			wantErr: false,
		},
		{
			name:   "empty output value is valid",
			prompt: "{{outputs.task1}}",
			outputs: map[string]string{
				"task1": "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplateOutputs(tt.prompt, tt.outputs)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("expected error containing %q, got: %v", tt.wantErrContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestTemplateVarRegex tests the regex pattern matching directly.
func TestTemplateVarRegex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantMatches []struct {
			full     string
			taskName string
		}
	}{
		{
			name:        "no matches",
			input:       "plain text",
			wantMatches: nil,
		},
		{
			name:  "simple match",
			input: "{{outputs.task1}}",
			wantMatches: []struct {
				full     string
				taskName string
			}{
				{"{{outputs.task1}}", "task1"},
			},
		},
		{
			name:  "multiple matches",
			input: "{{outputs.task1}} and {{outputs.task2}}",
			wantMatches: []struct {
				full     string
				taskName string
			}{
				{"{{outputs.task1}}", "task1"},
				{"{{outputs.task2}}", "task2"},
			},
		},
		{
			name:  "task names with hyphens",
			input: "{{outputs.task-name}}",
			wantMatches: []struct {
				full     string
				taskName string
			}{
				{"{{outputs.task-name}}", "task-name"},
			},
		},
		{
			name:  "task names with underscores",
			input: "{{outputs.task_name}}",
			wantMatches: []struct {
				full     string
				taskName string
			}{
				{"{{outputs.task_name}}", "task_name"},
			},
		},
		{
			name:  "task names with numbers",
			input: "{{outputs.task123}}",
			wantMatches: []struct {
				full     string
				taskName string
			}{
				{"{{outputs.task123}}", "task123"},
			},
		},
		{
			name:        "invalid - no closing braces",
			input:       "{{outputs.task1",
			wantMatches: nil,
		},
		{
			name:        "invalid - spaces in template",
			input:       "{{ outputs.task1 }}",
			wantMatches: nil,
		},
		{
			name:        "invalid - wrong prefix",
			input:       "{{output.task1}}",
			wantMatches: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := templateVarRegex.FindAllStringSubmatch(tt.input, -1)

			if tt.wantMatches == nil {
				if len(matches) != 0 {
					t.Errorf("expected no matches, got %d: %v", len(matches), matches)
				}
				return
			}

			if len(matches) != len(tt.wantMatches) {
				t.Errorf("expected %d matches, got %d", len(tt.wantMatches), len(matches))
				return
			}

			for i, want := range tt.wantMatches {
				if matches[i][0] != want.full {
					t.Errorf("match %d: full match = %q, want %q", i, matches[i][0], want.full)
				}
				if matches[i][1] != want.taskName {
					t.Errorf("match %d: task name = %q, want %q", i, matches[i][1], want.taskName)
				}
			}
		})
	}
}

// TestExpandPrompt_EdgeCases tests edge cases and potential security issues.
func TestExpandPrompt_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		outputs map[string]string
		want    string
	}{
		{
			name:   "output contains template syntax",
			prompt: "{{outputs.task1}}",
			outputs: map[string]string{
				"task1": "{{outputs.task2}}",
			},
			want: "{{outputs.task2}}", // Should not recursively expand
		},
		{
			name:   "very long output",
			prompt: "Result: {{outputs.task1}}",
			outputs: map[string]string{
				"task1": strings.Repeat("A", 10000),
			},
			want: "Result: " + strings.Repeat("A", 10000),
		},
		{
			name:   "output with newlines and special chars",
			prompt: "{{outputs.task1}}",
			outputs: map[string]string{
				"task1": "Line1\nLine2\r\nLine3\tTabbed",
			},
			want: "Line1\nLine2\r\nLine3\tTabbed",
		},
		{
			name:   "unicode in output",
			prompt: "{{outputs.task1}}",
			outputs: map[string]string{
				"task1": "Hello 世界 🌍",
			},
			want: "Hello 世界 🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandPrompt(tt.prompt, tt.outputs)
			if got != tt.want {
				t.Errorf("ExpandPrompt() = %q, want %q", got, tt.want)
			}
		})
	}
}
