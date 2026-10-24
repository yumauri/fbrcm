package contract

import "github.com/yumauri/fbrcm/core/themesource"

func classifyThemeSource(problem Problem, source *themesource.Error) Problem {
	problem.Details = struct {
		Kind   string `json:"kind"`
		Source string `json:"source"`
	}{Kind: "validation", Source: "theme_source"}
	switch source.Kind {
	case "invalid":
		problem.Code, problem.Category = "argument.invalid", "argument"
	case "ambiguous":
		problem.Code, problem.Category = "theme.source_ambiguous", "argument"
		candidates := make([]SelectionCandidate, 0, len(source.Candidates))
		for _, ref := range source.Candidates {
			candidates = append(candidates, SelectionCandidate{Name: ref, ID: ref})
			problem.Remediation = append(problem.Remediation, Remediation{Description: "Select the revision " + ref, Strategy: "retry_with_arguments", Argv: []string{"fbrcm", "theme", "import", source.SourceURL, "--ref", ref}})
		}
		problem.Details = SelectionDetails{Kind: "selection", Resource: "theme_source", Query: source.SourceURL, Candidates: candidates}
	case "not_found":
		problem.Code, problem.Category = "theme.source_not_found", "not_found"
	case "auth":
		problem.Code, problem.Category = "auth.credentials_invalid", "auth"
	case "permission":
		problem.Code, problem.Category = "theme.source_permission_denied", "permission"
	case "rate_limited":
		problem.Code, problem.Category, problem.Retryable = "theme.source_rate_limited", "unavailable", true
	case "limit":
		problem.Code, problem.Category = "theme.source_limit", "validation"
	case "timeout":
		problem.Code, problem.Category, problem.Retryable = "network.timeout", "timeout", true
	case "unavailable":
		problem.Code, problem.Category, problem.Retryable = "network.unavailable", "unavailable", true
	default:
		problem.Code, problem.Category = "theme.source_invalid_response", "validation"
	}
	if source.HTTPStatus > 0 {
		problem.Details = struct {
			Kind         string `json:"kind"`
			Service      string `json:"service"`
			Operation    string `json:"operation"`
			HTTPStatus   int    `json:"http_status"`
			RemoteStatus string `json:"remote_status"`
			RemoteCode   string `json:"remote_code"`
			RetryAfterMS int64  `json:"retry_after_ms"`
		}{Kind: "remote_api", Service: source.Provider, Operation: source.Operation, HTTPStatus: source.HTTPStatus, RetryAfterMS: source.RetryAfterMS}
	}
	return problem
}
