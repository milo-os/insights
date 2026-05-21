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
	"sync"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	envOnce sync.Once
	envInst *cel.Env
	envErr  error
)

// Environment returns the shared CEL environment configured for evaluating
// expressions against Kubernetes resources.
func Environment() (*cel.Env, error) {
	envOnce.Do(func() {
		envInst, envErr = cel.NewEnv(
			cel.Declarations(
				decls.NewVar("object", decls.NewMapType(decls.String, decls.Dyn)),
			),
		)
	})
	return envInst, envErr
}

// CompileCondition compiles a CEL expression that returns a boolean value.
// The expression has access to 'object' which represents the Kubernetes resource.
func CompileCondition(expression string) (cel.Program, error) {
	env, err := Environment()
	if err != nil {
		return nil, fmt.Errorf("failed to get CEL environment: %w", err)
	}

	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("failed to compile CEL expression: %w", issues.Err())
	}

	// Check that the expression returns a boolean
	if ast.OutputType() != cel.BoolType {
		return nil, fmt.Errorf("condition must evaluate to true/false, but expression returns %s. Ensure your condition uses comparison operators (==, !=, <, >) or boolean logic (&&, ||, !)", ast.OutputType())
	}

	// Use OptOptimize to enable expression optimization for better performance
	program, err := env.Program(ast, cel.EvalOptions(cel.OptOptimize))
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL program: %w", err)
	}

	return program, nil
}

// CompileExpression compiles a CEL expression that can return any type.
// The expression has access to 'object' which represents the Kubernetes resource.
func CompileExpression(expression string) (cel.Program, error) {
	env, err := Environment()
	if err != nil {
		return nil, fmt.Errorf("failed to get CEL environment: %w", err)
	}

	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("failed to compile CEL expression: %w", issues.Err())
	}

	// Use OptOptimize to enable expression optimization for better performance
	program, err := env.Program(ast, cel.EvalOptions(cel.OptOptimize))
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL program: %w", err)
	}

	return program, nil
}

// EvaluateCondition evaluates a compiled CEL condition program against a resource.
// Returns true if the condition matches, false otherwise.
func EvaluateCondition(program cel.Program, obj *unstructured.Unstructured) (bool, error) {
	result, err := evaluate(program, obj)
	if err != nil {
		return false, err
	}

	boolVal, ok := result.Value().(bool)
	if !ok {
		return false, fmt.Errorf("condition did not return true/false at runtime (got %T). This may indicate a type mismatch in your expression", result.Value())
	}

	return boolVal, nil
}

// EvaluateExpression evaluates a compiled CEL expression against a resource.
// Returns the result as a string.
func EvaluateExpression(program cel.Program, obj *unstructured.Unstructured) (string, error) {
	result, err := evaluate(program, obj)
	if err != nil {
		return "", err
	}

	return valueToString(result), nil
}

func evaluate(program cel.Program, obj *unstructured.Unstructured) (ref.Val, error) {
	activation := map[string]interface{}{
		"object": obj.Object,
	}

	result, _, err := program.Eval(activation)
	if err != nil {
		return nil, fmt.Errorf("expression evaluation failed: %w. Check that all referenced fields exist on the resource (use 'object.field' syntax)", err)
	}

	return result, nil
}

func valueToString(val ref.Val) string {
	switch v := val.(type) {
	case types.String:
		return string(v)
	case types.Int:
		return fmt.Sprintf("%d", int64(v))
	case types.Double:
		return fmt.Sprintf("%g", float64(v))
	case types.Bool:
		if bool(v) {
			return "true"
		}
		return "false"
	case types.Null:
		return ""
	default:
		// For complex types, use the native value's string representation
		nativeVal := val.Value()
		if nativeVal == nil {
			return ""
		}
		return fmt.Sprintf("%v", nativeVal)
	}
}
