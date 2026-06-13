package filter

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	exprruntime "github.com/expr-lang/expr/vm/runtime"
	"github.com/itchyny/gojq"

	"github.com/yumauri/fbrcm/core/firebase"
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

const rootGroupLabel = "(root)"

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
	return exprStringOperator(params[0], params[1], strings.Contains)
}

// exprStartsWith reports whether a string expression value starts with a prefix.
func exprStartsWith(params ...any) (any, error) {
	return exprStringOperator(params[0], params[1], strings.HasPrefix)
}

// exprEndsWith reports whether a string expression value ends with a suffix.
func exprEndsWith(params ...any) (any, error) {
	return exprStringOperator(params[0], params[1], strings.HasSuffix)
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

func exprValuesEqual(left, right any) bool {
	if _, ok := left.(rootGroup); ok {
		return right == nil || isRootGroupLabel(right) || exprruntime.Equal(left, right)
	}
	if _, ok := right.(rootGroup); ok {
		return left == nil || isRootGroupLabel(left) || exprruntime.Equal(left, right)
	}
	if leftValues, ok := left.(anyValue); ok {
		return leftValues.equal(right)
	}
	if rightValues, ok := right.(anyValue); ok {
		return rightValues.equal(left)
	}
	if exprruntime.Equal(left, right) {
		return true
	}
	return exprValuesCoerceEqual(left, right) || exprValuesCoerceEqual(right, left)
}

// isRootGroupLabel reports is root group label and returns the resulting value or error.
func isRootGroupLabel(value any) bool {
	text, ok := value.(string)
	return ok && text == rootGroupLabel
}

func applyParameterScope(env expressionEnv, param firebase.RemoteConfigParam) expressionEnv {
	env.Default = defaultRemoteConfigValueForExpr(param.DefaultValue, param.ValueType)
	env.Value = anyRemoteConfigValuesForExpr(param)
	env.Conditionals = make(map[string]any, len(param.ConditionalValues))
	for name, value := range param.ConditionalValues {
		env.Conditionals[name] = remoteConfigValueForExpr(value, param.ValueType)
	}
	return env
}

func anyRemoteConfigValuesForExpr(param firebase.RemoteConfigParam) anyValue {
	out := anyValue{
		values:    make([]any, 0, len(param.ConditionalValues)+1),
		valueType: strings.ToUpper(strings.TrimSpace(param.ValueType)),
	}
	if param.DefaultValue != nil {
		out.values = append(out.values, defaultRemoteConfigValueForExpr(param.DefaultValue, param.ValueType))
	}
	for _, name := range sortedConditionalValueKeys(param.ConditionalValues) {
		out.values = append(out.values, remoteConfigValueForExpr(param.ConditionalValues[name], param.ValueType))
	}
	return out
}

func sortedConditionalValueKeys(items map[string]firebase.RemoteConfigValue) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func defaultRemoteConfigValueForExpr(value *firebase.RemoteConfigValue, valueType string) any {
	if value == nil {
		return nil
	}
	return remoteConfigValueForExpr(*value, valueType)
}

func remoteConfigValueForExpr(value firebase.RemoteConfigValue, valueType string) any {
	switch {
	case value.UseInAppDefault:
		return "<in-app default>"
	case len(value.PersonalizationValue) > 0:
		return "<personalization>"
	case len(value.RolloutValue) > 0:
		return "<rollout>"
	}

	raw := strings.TrimSpace(value.Value)
	switch strings.ToUpper(strings.TrimSpace(valueType)) {
	case "BOOLEAN":
		switch raw {
		case "true":
			return true
		case "false":
			return false
		}
	case "NUMBER":
		if number, ok := exprParseJSONNumber(raw); ok {
			return number
		}
	}
	return value.Value
}

// exprValuesCoerceEqual compares typed values with string literals for backwards compatibility.
func exprValuesCoerceEqual(left, right any) bool {
	text, ok := left.(string)
	if !ok {
		return false
	}
	text = strings.TrimSpace(text)
	switch typed := right.(type) {
	case bool:
		return (text == "true" && typed) || (text == "false" && !typed)
	case int:
		return exprStringNumberEqual(text, float64(typed))
	case int8:
		return exprStringNumberEqual(text, float64(typed))
	case int16:
		return exprStringNumberEqual(text, float64(typed))
	case int32:
		return exprStringNumberEqual(text, float64(typed))
	case int64:
		return exprStringNumberEqual(text, float64(typed))
	case uint:
		return exprStringNumberEqual(text, float64(typed))
	case uint8:
		return exprStringNumberEqual(text, float64(typed))
	case uint16:
		return exprStringNumberEqual(text, float64(typed))
	case uint32:
		return exprStringNumberEqual(text, float64(typed))
	case uint64:
		return exprStringNumberEqual(text, float64(typed))
	case float32:
		return exprStringNumberEqual(text, float64(typed))
	case float64:
		return exprStringNumberEqual(text, typed)
	default:
		return false
	}
}

// equal reports whether any contained value equals right.
func (v anyValue) equal(right any) bool {
	if rightValues, ok := right.(anyValue); ok {
		for _, leftValue := range v.values {
			for _, rightValue := range rightValues.values {
				if exprValuesEqual(leftValue, rightValue) {
					return true
				}
			}
		}
		return false
	}
	for _, value := range v.values {
		if exprValuesEqual(value, right) {
			return true
		}
	}
	return false
}

// compare reports whether any contained value satisfies the comparison.
func (v anyValue) compare(right any, compare func(any, any) bool) (bool, error) {
	if rightValues, ok := right.(anyValue); ok {
		for _, leftValue := range v.values {
			for _, rightValue := range rightValues.values {
				matched, err := exprCompareScalar(leftValue, rightValue, compare)
				if err != nil {
					continue
				}
				if matched {
					return true, nil
				}
			}
		}
		return false, nil
	}
	for _, value := range v.values {
		matched, err := exprCompareScalar(value, right, compare)
		if err != nil {
			continue
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

// exprStringOperator applies a string operator to any string value, skipping non-strings.
func exprStringOperator(left, right any, match func(string, string) bool) (bool, error) {
	pattern, ok := right.(string)
	if !ok {
		return false, nil
	}
	return exprStringPredicate(left, func(text string) bool {
		return match(text, pattern)
	}), nil
}

// exprStringPredicate applies a string predicate to any string value, skipping non-strings.
func exprStringPredicate(left any, match func(string) bool) bool {
	if values, ok := left.(anyValue); ok {
		for _, value := range values.values {
			text, ok := value.(string)
			if ok && match(text) {
				return true
			}
		}
		return false
	}
	text, ok := left.(string)
	return ok && match(text)
}

// exprValueTypeMatches reports whether value matches a Firebase value type.
func exprValueTypeMatches(value any, valueType string) bool {
	if values, ok := value.(anyValue); ok {
		return strings.EqualFold(values.valueType, valueType)
	}
	switch strings.ToUpper(strings.TrimSpace(valueType)) {
	case "NUMBER":
		return exprIsNumberScalar(value)
	case "STRING", "JSON":
		_, ok := value.(string)
		return ok
	case "BOOLEAN":
		_, ok := value.(bool)
		return ok
	default:
		return false
	}
}

// exprIsNumberScalar reports whether value is a numeric scalar.
func exprIsNumberScalar(value any) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	default:
		return false
	}
}

// exprValueIsEmpty reports whether value is empty.
func exprValueIsEmpty(value any) bool {
	if values, ok := value.(anyValue); ok {
		return len(values.values) == 0 || slices.ContainsFunc(values.values, exprValueIsEmpty)
	}
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(typed) == ""
	case []any:
		return len(typed) == 0
	case map[string]any:
		return len(typed) == 0
	default:
		return false
	}
}

func exprValuesCompare(left, right any, compare func(any, any) bool) (bool, error) {
	if leftValues, ok := left.(anyValue); ok {
		return leftValues.compare(right, compare)
	}
	if rightValues, ok := right.(anyValue); ok {
		for _, rightValue := range rightValues.values {
			matched, err := exprCompareScalar(left, rightValue, compare)
			if err != nil {
				continue
			}
			if matched {
				return true, nil
			}
		}
		return false, nil
	}
	matched, err := exprCompareScalar(left, right, compare)
	if err != nil {
		return false, nil
	}
	return matched, nil
}

// exprCompareScalar compares two scalar values and preserves expr runtime errors.
func exprCompareScalar(left, right any, compare func(any, any) bool) (matched bool, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%v", recovered)
		}
	}()
	return compare(left, right), nil
}

// exprStringNumberEqual compares a string number to a typed number.
func exprStringNumberEqual(text string, right float64) bool {
	left, ok := exprParseJSONNumber(text)
	return ok && left == right
}

// exprParseJSONNumber parses Firebase NUMBER values without accepting NaN or Infinity.
func exprParseJSONNumber(text string) (float64, bool) {
	var number float64
	if err := json.Unmarshal([]byte(text), &number); err != nil {
		return 0, false
	}
	if math.IsInf(number, 0) || math.IsNaN(number) {
		return 0, false
	}
	return number, true
}

// compileJQ compiles and caches a gojq query.
func compileJQ(query string) (*gojq.Code, error) {
	if cached, ok := jqCodeCache.Load(query); ok {
		if err, ok := cached.(error); ok {
			return nil, err
		}
		return cached.(*gojq.Code), nil
	}
	parsed, err := gojq.Parse(query)
	if err != nil {
		jqCodeCache.Store(query, err)
		return nil, err
	}
	code, err := gojq.Compile(parsed)
	if err != nil {
		jqCodeCache.Store(query, err)
		return nil, err
	}
	jqCodeCache.Store(query, code)
	return code, nil
}

// jqResultsForValue runs code against value and returns successful results.
func jqResultsForValue(value any, code *gojq.Code) []any {
	if values, ok := value.(anyValue); ok {
		results := make([]any, 0, len(values.values))
		for _, item := range values.values {
			results = append(results, jqResultsForValue(item, code)...)
		}
		return results
	}
	input, ok := jqInputForValue(value)
	if !ok {
		return nil
	}
	results := []any{}
	iter := code.Run(input)
	for {
		result, ok := iter.Next()
		if !ok {
			return results
		}
		if _, ok := result.(error); ok {
			return results
		}
		results = append(results, result)
	}
}

// jqInputForValue converts an expression value into gojq input.
func jqInputForValue(value any) (any, bool) {
	text, ok := value.(string)
	if !ok {
		return value, true
	}
	var input any
	if err := json.Unmarshal([]byte(text), &input); err != nil {
		return nil, false
	}
	return input, true
}

// jqResultsAreBool reports whether all results are booleans.
func jqResultsAreBool(results []any) bool {
	for _, result := range results {
		if _, ok := result.(bool); !ok {
			return false
		}
	}
	return true
}

// prepareJQExpressions wraps unquoted jq(...) arguments as string literals.
func prepareJQExpressions(raw string) string {
	var out strings.Builder
	for i := 0; i < len(raw); {
		if isExprStringStart(raw[i]) {
			next := copyExprString(&out, raw, i)
			i = next
			continue
		}
		jqStart, paren := findJQCall(raw, i)
		if jqStart != i {
			out.WriteByte(raw[i])
			i++
			continue
		}
		closeParen := findMatchingParen(raw, paren)
		if closeParen < 0 {
			out.WriteString(raw[i:])
			break
		}
		out.WriteString(raw[jqStart : paren+1])
		arg := strings.TrimSpace(raw[paren+1 : closeParen])
		if arg == "" || isExprStringStart(arg[0]) {
			out.WriteString(raw[paren+1 : closeParen])
		} else {
			encoded, _ := json.Marshal(arg)
			out.Write(encoded)
		}
		out.WriteByte(')')
		i = closeParen + 1
	}
	return out.String()
}

// findJQCall reports whether a jq call starts at pos and returns its opening parenthesis.
func findJQCall(raw string, pos int) (int, int) {
	if pos > 0 && isIdentifierRune(rune(raw[pos-1])) {
		return -1, -1
	}
	if !strings.HasPrefix(raw[pos:], "jq") {
		return -1, -1
	}
	next := pos + len("jq")
	if next < len(raw) && isIdentifierRune(rune(raw[next])) {
		return -1, -1
	}
	for next < len(raw) && unicode.IsSpace(rune(raw[next])) {
		next++
	}
	if next >= len(raw) || raw[next] != '(' {
		return -1, -1
	}
	return pos, next
}

// findMatchingParen finds the closing parenthesis for open.
func findMatchingParen(raw string, open int) int {
	depth := 0
	for i := open; i < len(raw); {
		if isExprStringStart(raw[i]) {
			i = skipExprString(raw, i)
			continue
		}
		switch raw[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
		i++
	}
	return -1
}

// copyExprString copies a string literal and returns the next position.
func copyExprString(out *strings.Builder, raw string, start int) int {
	next := skipExprString(raw, start)
	out.WriteString(raw[start:next])
	return next
}

// skipExprString skips an expr or jq string literal.
func skipExprString(raw string, start int) int {
	quote := raw[start]
	for i := start + 1; i < len(raw); i++ {
		if raw[i] == '\\' {
			i++
			continue
		}
		if raw[i] == quote {
			return i + 1
		}
	}
	return len(raw)
}

// isExprStringStart reports whether b starts a string literal.
func isExprStringStart(b byte) bool {
	return b == '"' || b == '\'' || b == '`'
}

// isIdentifierRune reports whether r can be part of an identifier.
func isIdentifierRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
