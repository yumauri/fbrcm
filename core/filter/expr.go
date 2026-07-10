package filter

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	exprruntime "github.com/expr-lang/expr/vm/runtime"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rootgroup"
)

type Expression struct {
	raw     string
	program *vm.Program
}

type expressionEnv struct {
	ProjectID    string                  `expr:"project_id"`
	Project      string                  `expr:"project"`
	Conditions   []string                `expr:"conditions"`
	Groups       []string                `expr:"groups"`
	Parameters   map[string]parameterEnv `expr:"parameters"`
	Name         string                  `expr:"name"`
	Group        any                     `expr:"group"`
	Default      any                     `expr:"default"`
	Value        anyValue                `expr:"value"`
	Conditionals map[string]any          `expr:"conditionals"`
}

type parameterEnv struct {
	Group        any            `expr:"group"`
	Default      any            `expr:"default"`
	Value        anyValue       `expr:"value"`
	Conditionals map[string]any `expr:"conditionals"`
}

type rootGroup struct{}

// anyValue holds default and conditional values for expression matching.
type anyValue struct {
	values    []any
	valueType string
}

const rootGroupLabel = rootgroup.Label

var jqCodeCache sync.Map

func CompileExpression(raw string) (*Expression, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	prepared := prepareJQExpressions(raw)

	program, err := expr.Compile(
		prepared,
		expr.Env(expressionEnvTemplate()),
		expr.Function("fbrcm_expr_equal", exprEqual, func(any, any) bool { return false }),
		expr.Function("fbrcm_expr_not_equal", exprNotEqual, func(any, any) bool { return false }),
		expr.Function("fbrcm_expr_less", exprLess, func(any, any) bool { return false }),
		expr.Function("fbrcm_expr_less_or_equal", exprLessOrEqual, func(any, any) bool { return false }),
		expr.Function("fbrcm_expr_greater", exprGreater, func(any, any) bool { return false }),
		expr.Function("fbrcm_expr_greater_or_equal", exprGreaterOrEqual, func(any, any) bool { return false }),
		expr.Function("fbrcm_expr_contains", exprContains, func(any, any) bool { return false }),
		expr.Function("fbrcm_expr_starts_with", exprStartsWith, func(any, any) bool { return false }),
		expr.Function("fbrcm_expr_ends_with", exprEndsWith, func(any, any) bool { return false }),
		expr.Function("fbrcm_expr_matches", exprMatches, func(any, any) bool { return false }),
		expr.Function("is_number", exprIsNumber, func(any) bool { return false }),
		expr.Function("is_string", exprIsString, func(any) bool { return false }),
		expr.Function("is_json", exprIsJSON, func(any) bool { return false }),
		expr.Function("is_boolean", exprIsBoolean, func(any) bool { return false }),
		expr.Function("is_empty", exprIsEmpty, func(any) bool { return false }),
		expr.Function("jq", exprJQ, func(any, string) any { return false }),
		expr.Operator("==", "fbrcm_expr_equal"),
		expr.Operator("!=", "fbrcm_expr_not_equal"),
		expr.Operator("<", "fbrcm_expr_less"),
		expr.Operator("<=", "fbrcm_expr_less_or_equal"),
		expr.Operator(">", "fbrcm_expr_greater"),
		expr.Operator(">=", "fbrcm_expr_greater_or_equal"),
		expr.Operator("contains", "fbrcm_expr_contains"),
		expr.Operator("startsWith", "fbrcm_expr_starts_with"),
		expr.Operator("endsWith", "fbrcm_expr_ends_with"),
		expr.Operator("matches", "fbrcm_expr_matches"),
		expr.AsBool(),
	)
	if err != nil {
		return nil, err
	}

	return &Expression{
		raw:     raw,
		program: program,
	}, nil
}

func (e *Expression) MatchProject(projectID, projectName string, cfg *firebase.RemoteConfig) (bool, error) {
	return e.match(buildExpressionEnv(projectID, projectName, cfg, "", ""))
}

func (e *Expression) MatchParameter(projectID, projectName string, cfg *firebase.RemoteConfig, name, group string) (bool, error) {
	env := buildExpressionEnv(projectID, projectName, cfg, name, group)
	cfgGroup := remoteConfigGroupName(group)
	if cfg != nil {
		if param, ok := cfg.Parameters[name]; ok && cfgGroup == "" {
			env = applyParameterScope(env, param)
		} else if groupCfg, ok := cfg.ParameterGroups[cfgGroup]; ok {
			if param, ok := groupCfg.Parameters[name]; ok {
				env = applyParameterScope(env, param)
			}
		}
	}
	return e.match(env)
}

func (e *Expression) match(env expressionEnv) (bool, error) {
	if e == nil {
		return true, nil
	}

	out, err := expr.Run(e.program, env)
	if err != nil {
		return false, fmt.Errorf("run %q: %w", e.raw, err)
	}

	matched, ok := out.(bool)
	if !ok {
		return false, fmt.Errorf("run %q: expected bool result, got %T", e.raw, out)
	}

	return matched, nil
}

func exprEqual(params ...any) (any, error) {
	return exprValuesEqual(params[0], params[1]), nil
}

func exprNotEqual(params ...any) (any, error) {
	return !exprValuesEqual(params[0], params[1]), nil
}

func exprLess(params ...any) (any, error) {
	return exprValuesCompare(params[0], params[1], exprruntime.Less)
}

func exprLessOrEqual(params ...any) (any, error) {
	return exprValuesCompare(params[0], params[1], exprruntime.LessOrEqual)
}

func exprGreater(params ...any) (any, error) {
	return exprValuesCompare(params[0], params[1], exprruntime.More)
}

func exprGreaterOrEqual(params ...any) (any, error) {
	return exprValuesCompare(params[0], params[1], exprruntime.MoreOrEqual)
}

// exprContains reports whether a string expression value contains a substring.
func exprContains(params ...any) (any, error) {
	return exprStringOperator(params[0], params[1], strings.Contains), nil
}

// exprStartsWith reports whether a string expression value starts with a prefix.
func exprStartsWith(params ...any) (any, error) {
	return exprStringOperator(params[0], params[1], strings.HasPrefix), nil
}

// exprEndsWith reports whether a string expression value ends with a suffix.
func exprEndsWith(params ...any) (any, error) {
	return exprStringOperator(params[0], params[1], strings.HasSuffix), nil
}

// exprMatches reports whether a string expression value matches a regular expression.
func exprMatches(params ...any) (any, error) {
	pattern, ok := params[1].(string)
	if !ok {
		return false, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	return exprStringPredicate(params[0], re.MatchString), nil
}

// exprIsNumber reports whether value is a Firebase NUMBER expression value.
func exprIsNumber(params ...any) (any, error) {
	return exprValueTypeMatches(params[0], "NUMBER"), nil
}

// exprIsString reports whether value is a Firebase STRING expression value.
func exprIsString(params ...any) (any, error) {
	return exprValueTypeMatches(params[0], "STRING"), nil
}

// exprIsJSON reports whether value is a Firebase JSON expression value.
func exprIsJSON(params ...any) (any, error) {
	return exprValueTypeMatches(params[0], "JSON"), nil
}

// exprIsBoolean reports whether value is a Firebase BOOLEAN expression value.
func exprIsBoolean(params ...any) (any, error) {
	return exprValueTypeMatches(params[0], "BOOLEAN"), nil
}

// exprIsEmpty reports whether value is empty.
func exprIsEmpty(params ...any) (any, error) {
	return exprValueIsEmpty(params[0]), nil
}

// exprJQ runs a gojq query against an expression value.
func exprJQ(params ...any) (any, error) {
	if len(params) != 2 {
		return false, fmt.Errorf("jq expects value and query")
	}
	query, ok := params[1].(string)
	if !ok {
		return false, fmt.Errorf("jq query must be string")
	}
	code, err := compileJQ(query)
	if err != nil {
		return false, err
	}
	results := jqResultsForValue(params[0], code)
	if len(results) == 0 {
		return false, nil
	}
	if jqResultsAreBool(results) {
		return slices.Contains(results, true), nil
	}
	return anyValue{values: results}, nil
}
