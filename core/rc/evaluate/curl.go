package evaluate

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode"
)

func readCURL(input string) ([]Exchange, error) {
	words, err := curlWords(input)
	if err != nil {
		return nil, err
	}
	if len(words) == 0 || words[0] != "curl" {
		return nil, fmt.Errorf("expected a copied cURL command or a HAR archive")
	}
	var rawURL, body string
	headers := make(http.Header)
	method := "POST"
	for i := 1; i < len(words); i++ {
		word := words[i]
		if strings.HasPrefix(word, "https://") {
			if rawURL != "" {
				return nil, fmt.Errorf("expected one cURL request URL")
			}
			rawURL = word
			continue
		}
		name, value, attached := strings.Cut(word, "=")
		if !attached && len(word) > 2 && strings.Contains("HdXAb", word[1:2]) && word[0] == '-' && word[1] != '-' {
			name, value, attached = word[:2], word[2:], true
		}
		switch name {
		case "--compressed", "--globoff", "-g", "--silent", "-s", "--show-error", "-S", "--location", "-L", "--insecure", "-k", "--http1.1", "--http2":
			continue
		case "-H", "--header", "-A", "--user-agent", "-b", "--cookie", "-X", "--request", "--url", "-d", "--data", "--data-raw", "--data-binary", "--json":
			if !attached {
				i++
				if i >= len(words) {
					return nil, fmt.Errorf("cURL option %s requires a value", name)
				}
				value = words[i]
			}
		default:
			return nil, fmt.Errorf("unsupported cURL option; use an unmodified Copy as cURL export")
		}
		switch name {
		case "-H", "--header":
			if key, text, found := strings.Cut(value, ":"); found {
				recordCaptureHeader(headers, key, text)
			}
		case "-A", "--user-agent":
			recordCaptureHeader(headers, "User-Agent", value)
		case "--url":
			if rawURL != "" {
				return nil, fmt.Errorf("expected one cURL request URL")
			}
			rawURL = value
		case "-X", "--request":
			method = value
		case "-d", "--data", "--data-raw", "--data-binary", "--json":
			if body != "" || name != "--data-raw" && strings.HasPrefix(value, "@") {
				return nil, fmt.Errorf("cURL must contain one inline JSON body; external body files are not read")
			}
			body = value
		}
	}
	project, namespace, ok := fetchTarget(rawURL)
	if !ok || !strings.EqualFold(method, "POST") {
		return nil, fmt.Errorf("cURL must describe a Firebase Remote Config :fetch POST request")
	}
	decoded, err := decodeBody(harBody{Text: body})
	if err != nil {
		return nil, err
	}
	facts, err := requestFacts(decoded)
	if err != nil {
		return nil, fmt.Errorf("cURL has a missing or invalid JSON fetch body")
	}
	exchange := Exchange{Format: "curl", Entry: 1, Project: project, Namespace: namespace, Facts: facts}
	prepareExchange(&exchange, decoded, headers)
	return []Exchange{exchange}, nil
}

// curlWords understands the quoting emitted by browser Copy as cURL (including
// ANSI-C $'...' strings). It is a tokenizer, never a shell interpreter.
func curlWords(input string) ([]string, error) {
	var words []string
	var word strings.Builder
	active := false
	for i := 0; i < len(input); {
		c := input[i]
		switch {
		case unicode.IsSpace(rune(c)):
			if active {
				words = append(words, word.String())
				word.Reset()
				active = false
			}
			i++
		case c == '\\':
			if i+1 >= len(input) {
				return nil, fmt.Errorf("incomplete cURL escape")
			}
			if input[i+1] != '\n' {
				word.WriteByte(input[i+1])
				active = true
			}
			i += 2
		case c == '\'' || c == '"' || c == '$' && i+1 < len(input) && input[i+1] == '\'':
			ansiC := c == '$'
			if ansiC {
				i++
				c = '\''
			}
			i++
			active = true
			closed := false
			for i < len(input) {
				if input[i] == c {
					i++
					closed = true
					break
				}
				if ansiC && input[i] == '\\' {
					r, multibyte, tail, err := strconv.UnquoteChar(input[i:], '\'')
					if err != nil {
						return nil, fmt.Errorf("invalid ANSI-C cURL escape")
					}
					if multibyte {
						word.WriteRune(r)
					} else {
						word.WriteByte(byte(r))
					}
					i = len(input) - len(tail)
					continue
				}
				if c == '"' && input[i] == '\\' && i+1 < len(input) && strings.ContainsRune("$`\"\\\n", rune(input[i+1])) {
					i++
					if input[i] != '\n' {
						word.WriteByte(input[i])
					}
					i++
					continue
				}
				if c == '"' && (input[i] == '$' || input[i] == '`') {
					return nil, fmt.Errorf("shell expansion is not supported in captured cURL")
				}
				word.WriteByte(input[i])
				i++
			}
			if !closed {
				return nil, fmt.Errorf("unterminated cURL quote")
			}
		case strings.ContainsRune("$`|;&<>()", rune(c)):
			return nil, fmt.Errorf("shell syntax is not supported in captured cURL")
		default:
			active = true
			word.WriteByte(c)
			i++
		}
	}
	if active {
		words = append(words, word.String())
	}
	return words, nil
}
