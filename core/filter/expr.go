package filter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	exprruntime "github.com/expr-lang/expr/vm/runtime"

	"github.com/yumauri/fbrcm/core/firebase"
)

// Expression holds expression state used by the filter package.
type Expression struct {
	// raw stores raw for Expression.
	raw string
	// program stores program for Expression.
	program *vm.Program
}

// expressionEnv holds expression env state used by the filter package.
type expressionEnv struct {
	// ProjectID stores project id for expressionEnv.
	ProjectID string `expr:"project_id"`
	// Project stores project for expressionEnv.
	Project string `expr:"project"`
	// Conditions stores conditions for expressionEnv.
	Conditions []string `expr:"conditions"`
	// Groups stores groups for expressionEnv.
	Groups []string `expr:"groups"`
	// Parameters stores parameters for expressionEnv.
	Parameters map[string]parameterEnv `expr:"parameters"`
	// Name stores name for expressionEnv.
	Name string `expr:"name"`
	// Group stores group for expressionEnv.
	Group any `expr:"group"`
	// Default stores default for expressionEnv.
	Default any `expr:"default"`
	// Value stores value for expressionEnv.
	Value any `expr:"value"`
	// Conditionals stores conditionals for expressionEnv.
	Conditionals map[string]string `expr:"conditionals"`
}

// parameterEnv holds parameter env state used by the filter package.
type parameterEnv struct {
	// Group stores group for parameterEnv.
	Group any `expr:"group"`
	// Default stores default for parameterEnv.
	Default any `expr:"default"`
	// Value stores value for parameterEnv.
	Value any `expr:"value"`
	// Conditionals stores conditionals for parameterEnv.
	Conditionals map[string]string `expr:"conditionals"`
}

// rootGroup holds root group state used by the filter package.
type rootGroup struct{}

const rootGroupLabel = "(root)"

// CompileExpression handles compile expression and returns the resulting value or error.
func CompileExpression(raw string) (*Expression, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	program, err := expr.Compile(
		raw,
		expr.Env(expressionEnvTemplate()),
		expr.Function("fbrcm_expr_equal", exprEqual, func(any, any) bool { return false }),
		expr.Function("fbrcm_expr_not_equal", exprNotEqual, func(any, any) bool { return false }),
		expr.Operator("==", "fbrcm_expr_equal"),
		expr.Operator("!=", "fbrcm_expr_not_equal"),
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

// CompileProjectExpression handles compile project expression and returns the resulting value or error.
func CompileProjectExpression(raw string) (*Expression, error) {
	return CompileExpression(raw)
}

// MatchProject matches project for Expression and returns the resulting state or error.
func (e *Expression) MatchProject(projectID, projectName string, cfg *firebase.RemoteConfig) (bool, error) {
	return e.match(buildExpressionEnv(projectID, projectName, cfg, "", ""))
}

// MatchParameter matches parameter for Expression and returns the resulting state or error.
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

// match matches match for Expression and returns the resulting state or error.
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

// ProjectExpressionEnv handles project expression env and returns the resulting value or error.
func ProjectExpressionEnv(projectID, projectName string, cfg *firebase.RemoteConfig) expressionEnv {
	return buildExpressionEnv(projectID, projectName, cfg, "", "")
}

// buildExpressionEnv handles build expression env and returns the resulting value or error.
func buildExpressionEnv(projectID, projectName string, cfg *firebase.RemoteConfig, name, group string) expressionEnv {
	env := expressionEnv{
		ProjectID:    projectID,
		Project:      projectName,
		Conditions:   []string{},
		Groups:       []string{},
		Parameters:   map[string]parameterEnv{},
		Name:         name,
		Group:        groupValueForExpr(group),
		Conditionals: map[string]string{},
	}
	if cfg == nil {
		return env
	}

	conditions := make([]string, 0, len(cfg.Conditions))
	for _, condition := range cfg.Conditions {
		name := strings.TrimSpace(condition.Name)
		if name == "" {
			continue
		}
		conditions = append(conditions, name)
	}
	sort.Strings(conditions)
	env.Conditions = conditions

	groups := make([]string, 0, len(cfg.ParameterGroups))
	for groupName := range cfg.ParameterGroups {
		groupName = strings.TrimSpace(groupName)
		if groupName == "" {
			continue
		}
		groups = append(groups, groupName)
	}
	sort.Strings(groups)
	env.Groups = groups

	parameters := make(map[string]parameterEnv, len(cfg.Parameters)+len(cfg.ParameterGroups))
	for key, param := range cfg.Parameters {
		parameters[key] = parameterExpressionEnv("", param)
	}
	for groupName, group := range cfg.ParameterGroups {
		for key, param := range group.Parameters {
			parameters[key] = parameterExpressionEnv(groupName, param)
		}
	}
	env.Parameters = parameters

	return env
}

// expressionEnvTemplate handles expression env template and returns the resulting value or error.
func expressionEnvTemplate() expressionEnv {
	return expressionEnv{
		Conditions:   []string{},
		Groups:       []string{},
		Parameters:   map[string]parameterEnv{},
		Conditionals: map[string]string{},
	}
}

// parameterExpressionEnv handles parameter expression env and returns the resulting value or error.
func parameterExpressionEnv(groupName string, param firebase.RemoteConfigParam) parameterEnv {
	conditionals := make(map[string]string, len(param.ConditionalValues))
	for name, value := range param.ConditionalValues {
		conditionals[name] = remoteConfigValueForExpr(value)
	}

	return parameterEnv{
		Group:        groupValueForExpr(groupName),
		Default:      defaultRemoteConfigValueForExpr(param.DefaultValue),
		Value:        defaultRemoteConfigValueForExpr(param.DefaultValue),
		Conditionals: conditionals,
	}
}

// groupValueForExpr handles group value for expr and returns the resulting value or error.
func groupValueForExpr(groupName string) any {
	groupName = strings.TrimSpace(groupName)
	if groupName == "" || groupName == rootGroupLabel {
		return rootGroup{}
	}
	return groupName
}

// remoteConfigGroupName handles remote config group name and returns the resulting value or error.
func remoteConfigGroupName(groupName string) string {
	if strings.TrimSpace(groupName) == rootGroupLabel {
		return ""
	}
	return groupName
}

// exprEqual handles expr equal and returns the resulting value or error.
func exprEqual(params ...any) (any, error) {
	return exprValuesEqual(params[0], params[1]), nil
}

// exprNotEqual handles expr not equal and returns the resulting value or error.
func exprNotEqual(params ...any) (any, error) {
	return !exprValuesEqual(params[0], params[1]), nil
}

// exprValuesEqual handles expr values equal and returns the resulting value or error.
func exprValuesEqual(left, right any) bool {
	if _, ok := left.(rootGroup); ok {
		return right == nil || isRootGroupLabel(right) || exprruntime.Equal(left, right)
	}
	if _, ok := right.(rootGroup); ok {
		return left == nil || isRootGroupLabel(left) || exprruntime.Equal(left, right)
	}
	return exprruntime.Equal(left, right)
}

// isRootGroupLabel reports is root group label and returns the resulting value or error.
func isRootGroupLabel(value any) bool {
	text, ok := value.(string)
	return ok && text == rootGroupLabel
}

// applyParameterScope handles apply parameter scope and returns the resulting value or error.
func applyParameterScope(env expressionEnv, param firebase.RemoteConfigParam) expressionEnv {
	env.Default = defaultRemoteConfigValueForExpr(param.DefaultValue)
	env.Value = defaultRemoteConfigValueForExpr(param.DefaultValue)
	env.Conditionals = make(map[string]string, len(param.ConditionalValues))
	for name, value := range param.ConditionalValues {
		env.Conditionals[name] = remoteConfigValueForExpr(value)
	}
	return env
}

// defaultRemoteConfigValueForExpr handles default remote config value for expr and returns the resulting value or error.
func defaultRemoteConfigValueForExpr(value *firebase.RemoteConfigValue) any {
	if value == nil {
		return nil
	}
	return remoteConfigValueForExpr(*value)
}

// remoteConfigValueForExpr handles remote config value for expr and returns the resulting value or error.
func remoteConfigValueForExpr(value firebase.RemoteConfigValue) string {
	switch {
	case value.UseInAppDefault:
		return "<in-app default>"
	case len(value.PersonalizationValue) > 0:
		return "<personalization>"
	case len(value.RolloutValue) > 0:
		return "<rollout>"
	default:
		return value.Value
	}
}
