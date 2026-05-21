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
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestCompileCondition(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		wantErr    bool
	}{
		{
			name:       "valid boolean expression",
			expression: `object.metadata.name == "test"`,
			wantErr:    false,
		},
		{
			name:       "valid has expression",
			expression: `has(object.spec.replicas)`,
			wantErr:    false,
		},
		{
			name:       "invalid non-boolean return",
			expression: `object.metadata.name`,
			wantErr:    true,
		},
		{
			name:       "invalid syntax",
			expression: `object.metadata.name ==`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CompileCondition(tt.expression)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompileCondition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEvaluateCondition(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "nginx",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"replicas": int64(3),
			},
		},
	}

	tests := []struct {
		name       string
		expression string
		want       bool
	}{
		{
			name:       "name equals",
			expression: `object.metadata.name == "nginx"`,
			want:       true,
		},
		{
			name:       "name not equals",
			expression: `object.metadata.name == "apache"`,
			want:       false,
		},
		{
			name:       "replicas check",
			expression: `object.spec.replicas == 3`,
			want:       true,
		},
		{
			name:       "has field",
			expression: `has(object.spec.replicas)`,
			want:       true,
		},
		{
			name:       "missing field",
			expression: `has(object.spec.selector)`,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := CompileCondition(tt.expression)
			if err != nil {
				t.Fatalf("CompileCondition() error = %v", err)
			}

			got, err := EvaluateCondition(program, obj)
			if err != nil {
				t.Fatalf("EvaluateCondition() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("EvaluateCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompileTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantErr  bool
	}{
		{
			name:     "static text only",
			template: "This is static text",
			wantErr:  false,
		},
		{
			name:     "single expression",
			template: "Name: {{ object.metadata.name }}",
			wantErr:  false,
		},
		{
			name:     "multiple expressions",
			template: "{{ object.metadata.namespace }}/{{ object.metadata.name }}",
			wantErr:  false,
		},
		{
			name:     "expression with spaces",
			template: "Value: {{   object.metadata.name   }}",
			wantErr:  false,
		},
		{
			name:     "complex expression",
			template: `Status: {{ has(object.spec.replicas) ? "configured" : "not configured" }}`,
			wantErr:  false,
		},
		{
			name:     "invalid expression",
			template: "Bad: {{ object.metadata. }}",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CompileTemplate(tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompileTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCompiledTemplateEvaluate(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "nginx",
				"namespace": "default",
				"labels": map[string]interface{}{
					"app": "web",
				},
			},
			"spec": map[string]interface{}{
				"replicas": int64(3),
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name":  "nginx",
								"image": "nginx:1.21",
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name:     "static text",
			template: "Hello, World!",
			want:     "Hello, World!",
		},
		{
			name:     "single expression",
			template: "Deployment: {{ object.metadata.name }}",
			want:     "Deployment: nginx",
		},
		{
			name:     "namespace/name path",
			template: "{{ object.metadata.namespace }}/{{ object.metadata.name }}",
			want:     "default/nginx",
		},
		{
			name:     "integer value",
			template: "Replicas: {{ object.spec.replicas }}",
			want:     "Replicas: 3",
		},
		{
			name:     "ternary expression",
			template: `{{ has(object.spec.replicas) ? "Has replicas" : "No replicas" }}`,
			want:     "Has replicas",
		},
		{
			name:     "string concatenation",
			template: `{{ object.metadata.namespace + "/" + object.metadata.name }}`,
			want:     "default/nginx",
		},
		{
			name:     "mixed content",
			template: "The deployment {{ object.metadata.name }} in {{ object.metadata.namespace }} has {{ object.spec.replicas }} replicas.",
			want:     "The deployment nginx in default has 3 replicas.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, err := CompileTemplate(tt.template)
			if err != nil {
				t.Fatalf("CompileTemplate() error = %v", err)
			}

			got, err := ct.Evaluate(obj)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("Evaluate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTemplateEvaluator(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      "test",
				"namespace": "default",
			},
		},
	}

	te := NewTemplateEvaluator()

	// First evaluation should compile and cache
	result1, err := te.Evaluate("Name: {{ object.metadata.name }}", obj)
	if err != nil {
		t.Fatalf("First Evaluate() error = %v", err)
	}
	if result1 != "Name: test" {
		t.Errorf("First Evaluate() = %q, want %q", result1, "Name: test")
	}

	// Second evaluation should use cache
	result2, err := te.Evaluate("Name: {{ object.metadata.name }}", obj)
	if err != nil {
		t.Fatalf("Second Evaluate() error = %v", err)
	}
	if result2 != "Name: test" {
		t.Errorf("Second Evaluate() = %q, want %q", result2, "Name: test")
	}

	// Verify cache has one entry
	if len(te.cache) != 1 {
		t.Errorf("Cache size = %d, want 1", len(te.cache))
	}

	// Clear cache
	te.Clear()
	if len(te.cache) != 0 {
		t.Errorf("After Clear(), cache size = %d, want 0", len(te.cache))
	}
}
