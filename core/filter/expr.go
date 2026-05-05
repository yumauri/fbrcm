package filter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"fbrcm/core/firebase"
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
	Group        string                  `expr:"group"`
	Default      any                     `expr:"default"`
	Conditionals map[string]string       `expr:"conditionals"`
}

type parameterEnv struct {
	Group        string            `expr:"group"`
	Default      any               `expr:"default"`
	Conditionals map[string]string `expr:"conditionals"`
}

func CompileExpression(raw string) (*Expression, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	program, err := expr.Compile(
		raw,
		expr.Env(expressionEnvTemplate()),
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

func CompileProjectExpression(raw string) (*Expression, error) {
	return CompileExpression(raw)
}

func (e *Expression) MatchProject(projectID, projectName string, cfg *firebase.RemoteConfig) (bool, error) {
	return e.match(buildExpressionEnv(projectID, projectName, cfg, "", ""))
}

func (e *Expression) MatchParameter(projectID, projectName string, cfg *firebase.RemoteConfig, name, group string) (bool, error) {
	env := buildExpressionEnv(projectID, projectName, cfg, name, group)
	if cfg != nil {
		if param, ok := cfg.Parameters[name]; ok && group == "" {
			env = applyParameterScope(env, param)
		} else if groupCfg, ok := cfg.ParameterGroups[group]; ok {
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

func ProjectExpressionEnv(projectID, projectName string, cfg *firebase.RemoteConfig) expressionEnv {
	return buildExpressionEnv(projectID, projectName, cfg, "", "")
}

func buildExpressionEnv(projectID, projectName string, cfg *firebase.RemoteConfig, name, group string) expressionEnv {
	env := expressionEnv{
		ProjectID:    projectID,
		Project:      projectName,
		Conditions:   []string{},
		Groups:       []string{},
		Parameters:   map[string]parameterEnv{},
		Name:         name,
		Group:        group,
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

func expressionEnvTemplate() expressionEnv {
	return expressionEnv{
		Conditions:   []string{},
		Groups:       []string{},
		Parameters:   map[string]parameterEnv{},
		Conditionals: map[string]string{},
	}
}

func parameterExpressionEnv(groupName string, param firebase.RemoteConfigParam) parameterEnv {
	conditionals := make(map[string]string, len(param.ConditionalValues))
	for name, value := range param.ConditionalValues {
		conditionals[name] = remoteConfigValueForExpr(value)
	}

	return parameterEnv{
		Group:        groupName,
		Default:      defaultRemoteConfigValueForExpr(param.DefaultValue),
		Conditionals: conditionals,
	}
}

func applyParameterScope(env expressionEnv, param firebase.RemoteConfigParam) expressionEnv {
	env.Default = defaultRemoteConfigValueForExpr(param.DefaultValue)
	env.Conditionals = make(map[string]string, len(param.ConditionalValues))
	for name, value := range param.ConditionalValues {
		env.Conditionals[name] = remoteConfigValueForExpr(value)
	}
	return env
}

func defaultRemoteConfigValueForExpr(value *firebase.RemoteConfigValue) any {
	if value == nil {
		return nil
	}
	return remoteConfigValueForExpr(*value)
}

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
