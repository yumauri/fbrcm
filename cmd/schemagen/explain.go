package main

func captureCandidateSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"entry", "project", "namespace", "captured_at", "app_id", "http_status"}, "properties": map[string]any{
		"entry": map[string]any{"type": "integer", "minimum": 1}, "project": map[string]any{"type": "string"}, "namespace": map[string]any{"enum": []string{"firebase", "firebase-server"}}, "captured_at": map[string]any{"type": []string{"string", "null"}, "format": "date-time"}, "app_id": map[string]any{"type": []string{"string", "null"}}, "http_status": map[string]any{"type": "integer", "minimum": 0, "maximum": 599},
	}}
}

func captureStdinSchema() map[string]any {
	// encoding/json accepts omitted and null struct fields as zero values.
	// Validate fetch-specific fields only after selecting a Firebase request;
	// unrelated HAR traffic may have arbitrary URLs, statuses and encodings.
	text := map[string]any{"type": []string{"string", "null"}}
	object := func(properties map[string]any) map[string]any {
		return map[string]any{"type": []string{"object", "null"}, "properties": properties}
	}
	header := object(map[string]any{"name": text, "value": text})
	headers := map[string]any{"type": []string{"array", "null"}, "items": header}
	body := object(map[string]any{"text": text, "encoding": text})
	request := object(map[string]any{"method": text, "url": text, "headers": headers, "postData": body})
	response := object(map[string]any{"status": map[string]any{"type": []string{"integer", "null"}}, "headers": headers, "content": body})
	entry := object(map[string]any{"request": request, "response": response, "startedDateTime": map[string]any{"type": []string{"string", "null"}, "description": "Decoded by Go time.Time.UnmarshalJSON at the capture_document runtime boundary"}, "time": map[string]any{"type": []string{"number", "null"}}})
	// Presence and case-insensitive field lookup belong to the annotated
	// decoder boundary, so alternate Go field spellings are not excluded here.
	har := map[string]any{"type": "object", "properties": map[string]any{"log": object(map[string]any{"entries": map[string]any{"type": []string{"array", "null"}, "items": entry}})}}
	result := map[string]any{"description": "A raw cURL command or HAR JSON text, or a decoded HAR object. Format is detected automatically. Capture bytes and expanded gzip bodies are limited to 32 MiB. Only Firebase :fetch POST entries are selected; original one-based HAR indices are retained. Requests are never executed.", "oneOf": []any{map[string]any{"type": "string", "minLength": 1, "maxLength": 33554432}, har}}
	addValidationRule(result, map[string]any{"operator": "capture_document", "parser": "evaluate.ReadCapture", "maximum_bytes": 33554432})
	return result
}

func applyExplainFlagSchema(schema map[string]any, name string) {
	switch name {
	case "project":
		delete(schema, "items")
		delete(schema, "pattern")
		schema["allOf"] = []any{map[string]any{"$ref": "#/$defs/target_positional_selector"}}
		addMatchingRule(schema, map[string]any{"operator": "explain_project_resolution"})
	case "entry":
		schema["minimum"] = 0
		schema["maximum"] = int64(9223372036854775807)
		schema["description"] = "Original one-based HAR entry number; 0 selects automatically only when there is one Firebase fetch"
		addMatchingRule(schema, map[string]any{"operator": "explain_capture_entry_resolution"})
	case "set":
		item := map[string]any{"type": "string", "pattern": `^[\s\v]*(?:app\.(?:id|version|build|firebaseInstallationId|firstOpenTimestamp)|device\.(?:os|country|language|category|dateTime)|dateTime|app\.(?:browserAndVersion|operatingSystemAndVersion)\.(?:name|version)|app\.(?:customSignal|userProperty|audiences)[\s\v]*\[[\s\v]*(?:'(?:\\.|[^'\\])*'|"(?:\\.|[^"\\])*")[\s\v]*\]|percent(?:[\s\v]*\[[\s\v]*(?:'(?:\\.|[^'\\])*'|"(?:\\.|[^"\\])*")[\s\v]*\])?)[\s\v]*=`}
		item["description"] = "field=value input override. Indexed keys accept single or double quoted names. Values remain strings; /unknown and /unset are markers, and JSON-quoted strings escape them. Audience membership accepts true/false, yes/no, y/n (case-insensitive). Percentiles must parse as a finite number in [0,100]. Timestamps accept RFC3339 or ISO date-time interpreted as UTC. dateTime aliases device.dateTime. Later assignments to the same canonical field win."
		addValidationRule(item, map[string]any{"operator": "evaluate_input_assignment", "parser": "evaluate.InputNeed+evaluate.Answer", "percentile_minimum": 0, "percentile_maximum": 100, "unknown_marker": "/unknown", "absent_marker": "/unset"})
		schema["items"] = item
	case "no-input":
		schema["description"] = "Accepted for parity with human mode; JSON mode always leaves unavailable inputs unknown without prompting"
	}
}

func explainInputConstraints() []any {
	// A project may come from the capture; otherwise it must be explicit.
	missingCapture := map[string]any{"allOf": []any{map[string]any{"properties": map[string]any{"stdin": map[string]any{"type": "null"}}}, map[string]any{"properties": map[string]any{"options": map[string]any{"not": map[string]any{"required": []string{"from"}}}}}}}
	return []any{map[string]any{"if": missingCapture, "then": map[string]any{"properties": map[string]any{"options": map[string]any{"required": []string{"project"}, "properties": map[string]any{"entry": map[string]any{"const": 0}}}}}}}
}
