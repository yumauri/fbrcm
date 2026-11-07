package evaluate

import (
	"fmt"
	"strings"
)

type clientVersionTarget struct {
	name    string
	version *rule
}

// Web browser and OS conditions contain typed name/version targets. The name
// is resolved first; an irrelevant target never asks for a version.
func evalClientVersion(tokens []token, ctx *Context) (State, string, error) {
	field := strings.TrimSuffix(tokens[0].text, ".inOne")
	targets, err := parseClientTargets(tokens)
	if err != nil {
		return Unknown, err.Error(), nil
	}
	name, err := ctx.fact(Need{Key: field + ".name", Label: field + " name", Kind: "text"})
	if err != nil {
		return Unknown, "", err
	}
	if !name.Known {
		if name.Reason != "" {
			return Unknown, name.Reason, nil
		}
		return Unknown, field + " name is unavailable", nil
	}
	if name.Absent {
		return False, field + " name is unset", nil
	}
	state := False
	note := fmt.Sprintf("name = %s (%s)", quote(name.Value), name.Source)
	var names, evidence []string
	for _, target := range targets {
		names = append(names, target.name)
		if name.Value != target.name {
			continue
		}
		if target.version == nil {
			return True, note + " exactly matches " + quote(target.name) + " and any version is accepted", nil
		}
		version, err := ctx.fact(Need{Key: field + ".version", Label: field + " version", Kind: "text"})
		if err != nil {
			return Unknown, "", err
		}
		if !version.Known {
			state = Unknown
			if version.Reason != "" {
				evidence = append(evidence, version.Reason)
			} else {
				evidence = append(evidence, "version is unavailable")
			}
			continue
		}
		if version.Absent {
			evidence = append(evidence, "version is unset ("+version.Source+")")
			continue
		}
		match, err := compareRule(*target.version, version.Value)
		if err != nil {
			state = Unknown
			evidence = append(evidence, err.Error())
			continue
		}
		evidence = append(evidence, fmt.Sprintf("version = %s (%s) %s", quote(version.Value), version.Source, describeRule(*target.version, match)))
		if match {
			return True, note + " exactly matches " + quote(target.name) + ", " + strings.Join(evidence, ", "), nil
		}
	}
	if len(evidence) == 0 {
		return state, note + " " + describeRule(rule{op: "in", values: names}, false), nil
	}
	return state, note + " exactly matches " + quote(name.Value) + ", " + strings.Join(evidence, ", "), nil
}

func parseClientTargets(tokens []token) ([]clientVersionTarget, error) {
	invalid := fmt.Errorf("unsupported browser or operating-system version target")
	if len(tokens) < 10 || tokens[1].text != "(" || tokens[2].text != "[" || tokens[len(tokens)-2].text != "]" || tokens[len(tokens)-1].text != ")" {
		return nil, invalid
	}
	constructor := "browserName"
	if strings.HasPrefix(tokens[0].text, "app.operatingSystemAndVersion") {
		constructor = "operatingSystemName"
	}
	var targets []clientVersionTarget
	for i := 3; i < len(tokens)-2; {
		if i+4 >= len(tokens)-2 || tokens[i].text != constructor || tokens[i+1].text != "(" || !tokens[i+2].quoted || tokens[i+3].text != ")" {
			return nil, invalid
		}
		target := clientVersionTarget{name: tokens[i+2].text}
		i += 4
		if tokens[i].text == ".anyVersion" {
			i++
		} else {
			start, depth := i, 0
			for i < len(tokens)-2 {
				if !tokens[i].quoted {
					if tokens[i].text == "," && depth == 0 {
						break
					}
					switch tokens[i].text {
					case "(", "[":
						depth++
					case ")", "]":
						depth--
					}
				}
				i++
			}
			version, err := parseClientVersion(tokens[start:i])
			if err != nil {
				return nil, invalid
			}
			target.version = &version
		}
		targets = append(targets, target)
		if i < len(tokens)-2 {
			if tokens[i].text != "," {
				return nil, invalid
			}
			i++
		}
	}
	return targets, nil
}

func parseClientVersion(tokens []token) (rule, error) {
	if len(tokens) >= 5 && tokens[0].text == ".version." && tokens[2].text == "(" && tokens[len(tokens)-1].text == ")" {
		values, err := comparisonArguments(tokens[2:])
		return rule{field: "app.version", op: tokens[1].text, values: values, numericMethod: true}, err
	}
	if len(tokens) >= 4 && strings.HasPrefix(tokens[0].text, ".version.") {
		rewritten := append([]token(nil), tokens...)
		rewritten[0].text = "app.version." + strings.TrimPrefix(tokens[0].text, ".version.")
		return parseRule(rewritten)
	}
	return rule{}, fmt.Errorf("unsupported version selector")
}
