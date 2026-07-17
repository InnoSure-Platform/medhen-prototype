package engine

import (
	"fmt"
	"sync"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Evaluator provides a thread-safe registry of compiled AST expressions.
type Evaluator struct {
	mu      sync.RWMutex
	program map[string]*vm.Program
}

// NewEvaluator initializes a new rules execution engine.
func NewEvaluator() *Evaluator {
	return &Evaluator{
		program: make(map[string]*vm.Program),
	}
}

// Register compiles and caches an expression.
func (e *Evaluator) Register(ruleID string, expression string, env interface{}) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	program, err := expr.Compile(expression, expr.Env(env))
	if err != nil {
		return fmt.Errorf("failed to compile expression %s: %w", ruleID, err)
	}

	e.program[ruleID] = program
	return nil
}

// EvaluateRule executes a compiled rule against the provided context.
func (e *Evaluator) EvaluateRule(ruleID string, env interface{}) (bool, error) {
	e.mu.RLock()
	program, exists := e.program[ruleID]
	e.mu.RUnlock()

	if !exists {
		return false, fmt.Errorf("rule %s not found in registry", ruleID)
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate rule %s: %w", ruleID, err)
	}

	result, ok := output.(bool)
	if !ok {
		return false, fmt.Errorf("rule %s did not return a boolean value", ruleID)
	}

	return result, nil
}
