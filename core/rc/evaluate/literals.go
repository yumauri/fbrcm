package evaluate

import "strconv"

// ValueSpan identifies a comparison literal without changing its original text.
type ValueSpan struct {
	Start, End int
	Value      string
}

// ComparisonValueSpans reuses the expression lexer for human display of rules
// and evaluation evidence. Indexed input names and percentile seeds identify
// inputs rather than comparison values. Unrecognized text has no spans.
func ComparisonValueSpans(text string) []ValueSpan {
	tokens, err := lex(text)
	if err != nil {
		return nil
	}
	var spans []ValueSpan
	for i, token := range tokens {
		if token.quoted && i >= 2 {
			field, delimiter := tokens[i-2].text, tokens[i-1].text
			if delimiter == "[" && (field == "app.customSignal" || field == "app.userProperty" || field == "percent") ||
				delimiter == "(" && field == "percent" {
				continue
			}
		}
		literal := token.quoted || token.text == "true" || token.text == "false"
		if !literal {
			_, err := strconv.ParseFloat(token.text, 64)
			literal = err == nil
		}
		if literal {
			spans = append(spans, ValueSpan{Start: token.start, End: token.end, Value: token.text})
		}
	}
	return spans
}
