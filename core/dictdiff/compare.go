package dictdiff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pmezard/go-difflib/difflib"

	"github.com/yumauri/fbrcm/core/strfold"
)

const defaultContextLines = 3

func Compare(input Input) (Result, error) {
	result := Result{
		EntityName: input.EntityName,
		LeftName:   input.Left.Name,
		RightName:  input.Right.Name,
	}
	contextLines := input.ContextLines
	if contextLines <= 0 {
		contextLines = defaultContextLines
	}

	names := propertyNames(input.Left.Properties, input.Right.Properties)
	for _, name := range names {
		leftValue, leftOK := input.Left.Properties[name]
		rightValue, rightOK := input.Right.Properties[name]
		left, err := prepareValue(name, "left", leftValue, leftOK)
		if err != nil {
			return Result{}, err
		}
		right, err := prepareValue(name, "right", rightValue, rightOK)
		if err != nil {
			return Result{}, err
		}
		if leftOK && rightOK && left.Type == right.Type && left.Text == right.Text {
			continue
		}

		property := Property{Name: name, Left: left, Right: right}
		switch {
		case !leftOK:
			property.Kind = ChangeAdded
		case !rightOK:
			property.Kind = ChangeRemoved
		default:
			property.Kind = ChangeChanged
		}
		if left != nil && right != nil && compareAtomically(left, right) {
			property.Chunks = atomicChunks(left, right)
		} else {
			property.Chunks = diffChunks(preparedLines(left), preparedLines(right), contextLines)
		}
		if len(property.Chunks) == 0 {
			leftSegments, rightSegments := inlineSegments(left.Text, right.Text)
			if left.Text == right.Text {
				leftSegments = []Segment{{Text: left.Text, Kind: LineRemoved}}
				rightSegments = []Segment{{Text: right.Text, Kind: LineAdded}}
			}
			property.Chunks = []Chunk{{Rows: []Row{{
				Left:  &Line{Text: left.Text, Segments: leftSegments},
				Right: &Line{Text: right.Text, Segments: rightSegments},
				Kind:  LineChanged,
			}}}}
		}
		result.Properties = append(result.Properties, property)
	}
	return result, nil
}

func propertyNames(left, right Dictionary) []string {
	names := make([]string, 0, len(left)+len(right))
	seen := make(map[string]struct{}, len(left)+len(right))
	for name := range left {
		names = append(names, name)
		seen[name] = struct{}{}
	}
	for name := range right {
		if _, ok := seen[name]; !ok {
			names = append(names, name)
		}
	}
	strfold.Sort(names)
	return names
}

func prepareValue(property, side string, value Value, present bool) (*PreparedValue, error) {
	if !present {
		return nil, nil
	}
	compareAs, err := comparisonHint(value)
	if err != nil {
		return nil, fmt.Errorf("%s property %q: %w", side, property, err)
	}
	prepared := &PreparedValue{Type: value.Type, CompareAs: compareAs}
	switch value.Type {
	case ValueString, ValueJSON:
		prepared.Text = value.Raw
	case ValueBoolean:
		prepared.Text = strconv.FormatBool(value.Boolean)
	case ValueNumber:
		raw := strings.TrimSpace(value.Raw)
		decoder := json.NewDecoder(strings.NewReader(raw))
		decoder.UseNumber()
		var decoded any
		if !json.Valid([]byte(raw)) || decoder.Decode(&decoded) != nil {
			return nil, fmt.Errorf("%s property %q has invalid number %q", side, property, value.Raw)
		}
		number, ok := decoded.(json.Number)
		if !ok {
			return nil, fmt.Errorf("%s property %q has invalid number %q", side, property, value.Raw)
		}
		prepared.Text = number.String()
	case ValueNull:
		prepared.Text = "null"
	default:
		return nil, fmt.Errorf("%s property %q has unsupported value type %q", side, property, value.Type)
	}
	if compareAs == CompareJSON {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, []byte(prepared.Text), "", "  "); err == nil {
			prepared.Text = pretty.String()
		}
	}
	return prepared, nil
}

func comparisonHint(value Value) (ComparisonHint, error) {
	if value.CompareAs != "" {
		switch value.CompareAs {
		case CompareString, CompareJSON, CompareEnum:
			return value.CompareAs, nil
		default:
			return "", fmt.Errorf("unsupported comparison hint %q", value.CompareAs)
		}
	}
	switch value.Type {
	case ValueString:
		return CompareString, nil
	case ValueJSON:
		return CompareJSON, nil
	case ValueBoolean, ValueNumber, ValueNull:
		return CompareEnum, nil
	default:
		return "", nil
	}
}

func compareAtomically(left, right *PreparedValue) bool {
	return left.CompareAs == CompareEnum ||
		right.CompareAs == CompareEnum ||
		left.Type != right.Type
}

func atomicChunks(left, right *PreparedValue) []Chunk {
	leftLines := preparedLines(left)
	rightLines := preparedLines(right)
	rows := make([]Row, 0, max(len(leftLines), len(rightLines)))
	for index := range max(len(leftLines), len(rightLines)) {
		var leftLine, rightLine *Line
		if index < len(leftLines) {
			leftLine = atomicLine(leftLines[index])
		}
		if index < len(rightLines) {
			rightLine = atomicLine(rightLines[index])
		}
		rows = append(rows, Row{Left: leftLine, Right: rightLine, Kind: LineChanged})
	}
	return []Chunk{{Rows: rows}}
}

func atomicLine(text string) *Line {
	return &Line{
		Text:     text,
		Segments: []Segment{{Text: text, Kind: LineChanged}},
	}
}

func preparedLines(value *PreparedValue) []string {
	if value == nil {
		return nil
	}
	return strings.Split(value.Text, "\n")
}

func diffChunks(left, right []string, context int) []Chunk {
	opcodes := difflib.NewMatcher(left, right).GetOpCodes()
	chunks := make([]Chunk, 0)
	for index, opcode := range opcodes {
		if opcode.Tag == 'e' {
			continue
		}
		rows := make([]Row, 0)
		if index > 0 && opcodes[index-1].Tag == 'e' {
			previous := opcodes[index-1]
			start := max(previous.I2-context, previous.I1)
			for line := start; line < previous.I2; line++ {
				rows = append(rows, equalRow(left[line]))
			}
		}
		rows = append(rows, changedRows(opcode, left, right)...)
		if index+1 < len(opcodes) && opcodes[index+1].Tag == 'e' {
			next := opcodes[index+1]
			end := min(next.I1+context, next.I2)
			for line := next.I1; line < end; line++ {
				rows = append(rows, equalRow(left[line]))
			}
		}
		chunks = append(chunks, Chunk{Rows: rows})
	}
	return chunks
}

func changedRows(opcode difflib.OpCode, left, right []string) []Row {
	leftCount := opcode.I2 - opcode.I1
	rightCount := opcode.J2 - opcode.J1
	count := max(leftCount, rightCount)
	rows := make([]Row, 0, count)
	for offset := range count {
		var leftLine, rightLine *Line
		if offset < leftCount {
			leftLine = &Line{Text: left[opcode.I1+offset]}
		}
		if offset < rightCount {
			rightLine = &Line{Text: right[opcode.J1+offset]}
		}
		kind := LineChanged
		if leftLine == nil {
			kind = LineAdded
		} else if rightLine == nil {
			kind = LineRemoved
		} else {
			leftLine.Segments, rightLine.Segments = inlineSegments(leftLine.Text, rightLine.Text)
		}
		rows = append(rows, Row{Left: leftLine, Right: rightLine, Kind: kind})
	}
	return rows
}

func inlineSegments(left, right string) ([]Segment, []Segment) {
	leftRunes, rightRunes := runeStrings(left), runeStrings(right)
	opcodes := difflib.NewMatcher(leftRunes, rightRunes).GetOpCodes()
	leftSegments := make([]Segment, 0, len(opcodes))
	rightSegments := make([]Segment, 0, len(opcodes))
	for _, opcode := range opcodes {
		switch opcode.Tag {
		case 'e':
			text := strings.Join(leftRunes[opcode.I1:opcode.I2], "")
			leftSegments = appendSegment(leftSegments, Segment{Text: text, Kind: LineEqual})
			rightSegments = appendSegment(rightSegments, Segment{Text: text, Kind: LineEqual})
		case 'd':
			text := strings.Join(leftRunes[opcode.I1:opcode.I2], "")
			leftSegments = appendSegment(leftSegments, Segment{Text: text, Kind: LineRemoved})
		case 'i':
			text := strings.Join(rightRunes[opcode.J1:opcode.J2], "")
			rightSegments = appendSegment(rightSegments, Segment{Text: text, Kind: LineAdded})
		case 'r':
			leftText := strings.Join(leftRunes[opcode.I1:opcode.I2], "")
			rightText := strings.Join(rightRunes[opcode.J1:opcode.J2], "")
			leftSegments = appendSegment(leftSegments, Segment{Text: leftText, Kind: LineRemoved})
			rightSegments = appendSegment(rightSegments, Segment{Text: rightText, Kind: LineAdded})
		}
	}
	return leftSegments, rightSegments
}

func runeStrings(value string) []string {
	out := make([]string, 0, len(value))
	for _, r := range value {
		out = append(out, string(r))
	}
	return out
}

func appendSegment(segments []Segment, segment Segment) []Segment {
	if segment.Text == "" {
		return segments
	}
	if len(segments) > 0 && segments[len(segments)-1].Kind == segment.Kind {
		segments[len(segments)-1].Text += segment.Text
		return segments
	}
	return append(segments, segment)
}

func equalRow(text string) Row {
	return Row{
		Left:  &Line{Text: text},
		Right: &Line{Text: text},
		Kind:  LineEqual,
	}
}
