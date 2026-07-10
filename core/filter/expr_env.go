package filter

import (
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/strfold"
)

func buildExpressionEnv(projectID, projectName string, cfg *firebase.RemoteConfig, name, group string) expressionEnv {
	env := expressionEnv{
		ProjectID:    projectID,
		Project:      projectName,
		Conditions:   []string{},
		Groups:       []string{},
		Parameters:   map[string]parameterEnv{},
		Name:         name,
		Group:        groupValueForExpr(group),
		Conditionals: map[string]any{},
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
	strfold.Sort(conditions)
	env.Conditions = conditions

	groups := make([]string, 0, len(cfg.ParameterGroups))
	for groupName := range cfg.ParameterGroups {
		groupName = strings.TrimSpace(groupName)
		if groupName == "" {
			continue
		}
		groups = append(groups, groupName)
	}
	strfold.Sort(groups)
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
		Conditionals: map[string]any{},
	}
}

func parameterExpressionEnv(groupName string, param firebase.RemoteConfigParam) parameterEnv {
	conditionals := make(map[string]any, len(param.ConditionalValues))
	for name, value := range param.ConditionalValues {
		conditionals[name] = remoteConfigValueForExpr(value, param.ValueType)
	}

	return parameterEnv{
		Group:        groupValueForExpr(groupName),
		Default:      defaultRemoteConfigValueForExpr(param.DefaultValue, param.ValueType),
		Value:        anyRemoteConfigValuesForExpr(param),
		Conditionals: conditionals,
	}
}

func groupValueForExpr(groupName string) any {
	groupName = strings.TrimSpace(groupName)
	if groupName == "" || groupName == rootGroupLabel {
		return rootGroup{}
	}
	return groupName
}

func remoteConfigGroupName(groupName string) string {
	if strings.TrimSpace(groupName) == rootGroupLabel {
		return ""
	}
	return groupName
}
