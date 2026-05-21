/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cel

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/google/cel-go/cel"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// templatePattern matches {{ cel_expression }} placeholders in template strings.
var templatePattern = regexp.MustCompile(`\{\{\s*(.+?)\s*\}\}`)

// CompiledTemplate represents a parsed and compiled template with CEL expressions.
type CompiledTemplate struct {
	// Original template string
	original string
	// Parts of the template (alternating static strings and expression indices)
	parts []templatePart
	// Compiled CEL programs for each expression
	programs []cel.Program
}

type templatePart struct {
	// If isExpression is true, index points to the programs slice
	// If false, text contains the static string
	isExpression bool
	text         string
	index        int
	expression   string // Original expression text for error messages
}

// CompileTemplate parses a template string and compiles all CEL expressions.
// Template strings can contain {{ cel_expression }} placeholders.
func CompileTemplate(template string) (*CompiledTemplate, error) {
	ct := &CompiledTemplate{
		original: template,
	}

	// Find all matches and their positions
	matches := templatePattern.FindAllStringSubmatchIndex(template, -1)
	if len(matches) == 0 {
		// No expressions, just a static string
		ct.parts = []templatePart{{text: template}}
		return ct, nil
	}

	lastEnd := 0
	for _, match := range matches {
		// match[0] and match[1] are the full match bounds {{ ... }}
		// match[2] and match[3] are the capture group bounds (the expression)
		fullStart, fullEnd := match[0], match[1]
		exprStart, exprEnd := match[2], match[3]

		// Add static text before this expression
		if fullStart > lastEnd {
			ct.parts = append(ct.parts, templatePart{
				text: template[lastEnd:fullStart],
			})
		}

		// Extract and compile the CEL expression
		expression := strings.TrimSpace(template[exprStart:exprEnd])
		program, err := CompileExpression(expression)
		if err != nil {
			return nil, fmt.Errorf("failed to compile template expression %q: %w", expression, err)
		}

		ct.parts = append(ct.parts, templatePart{
			isExpression: true,
			index:        len(ct.programs),
			expression:   expression,
		})
		ct.programs = append(ct.programs, program)

		lastEnd = fullEnd
	}

	// Add any remaining static text
	if lastEnd < len(template) {
		ct.parts = append(ct.parts, templatePart{
			text: template[lastEnd:],
		})
	}

	return ct, nil
}

// Evaluate evaluates the compiled template against a Kubernetes resource.
// All {{ cel_expression }} placeholders are replaced with their evaluated values.
func (ct *CompiledTemplate) Evaluate(obj *unstructured.Unstructured) (string, error) {
	if len(ct.programs) == 0 {
		// No expressions, return original template
		if len(ct.parts) == 1 {
			return ct.parts[0].text, nil
		}
		return ct.original, nil
	}

	var result strings.Builder
	for _, part := range ct.parts {
		if !part.isExpression {
			result.WriteString(part.text)
			continue
		}

		value, err := EvaluateExpression(ct.programs[part.index], obj)
		if err != nil {
			return "", fmt.Errorf("failed to evaluate template expression '{{ %s }}': %w", part.expression, err)
		}
		result.WriteString(value)
	}

	return result.String(), nil
}

// ValidateTemplate checks if a template string is valid without fully compiling it.
// Returns an error if any CEL expression in the template is invalid.
func ValidateTemplate(template string) error {
	_, err := CompileTemplate(template)
	return err
}

// TemplateEvaluator provides a convenient interface for evaluating templates.
// It is safe for concurrent use.
type TemplateEvaluator struct {
	mu    sync.RWMutex
	cache map[string]*CompiledTemplate
}

// NewTemplateEvaluator creates a new TemplateEvaluator with caching.
func NewTemplateEvaluator() *TemplateEvaluator {
	return &TemplateEvaluator{
		cache: make(map[string]*CompiledTemplate),
	}
}

// Evaluate evaluates a template string against a resource.
// Compiled templates are cached for reuse. This method is safe for concurrent use.
func (te *TemplateEvaluator) Evaluate(template string, obj *unstructured.Unstructured) (string, error) {
	// Try read lock first for cache hit
	te.mu.RLock()
	ct, ok := te.cache[template]
	te.mu.RUnlock()

	if !ok {
		// Cache miss - need to compile and store
		var err error
		ct, err = CompileTemplate(template)
		if err != nil {
			return "", err
		}

		te.mu.Lock()
		// Double-check in case another goroutine compiled it
		if existing, exists := te.cache[template]; exists {
			ct = existing
		} else {
			te.cache[template] = ct
		}
		te.mu.Unlock()
	}
	return ct.Evaluate(obj)
}

// Clear removes all cached templates.
func (te *TemplateEvaluator) Clear() {
	te.mu.Lock()
	defer te.mu.Unlock()
	te.cache = make(map[string]*CompiledTemplate)
}
