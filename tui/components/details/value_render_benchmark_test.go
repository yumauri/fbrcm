package details

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
)

func BenchmarkSetDataWithLargeValue(b *testing.B) {
	for _, test := range []struct {
		name      string
		value     string
		valueType string
	}{
		{name: "JSON", value: `{"items":["` + strings.Repeat("abcdefghij", 56_000) + `"]}`, valueType: "JSON"},
		{name: "multiline string", value: strings.Repeat("large multiline value\n", 10_000), valueType: "STRING"},
	} {
		b.Run(test.name, func(b *testing.B) {
			data := &messages.ParameterViewData{Parameter: core.ParametersEntry{
				Key: "huge",
				Values: []core.ParametersValue{{
					Label: "default", Value: test.value, RawValue: test.value, ValueType: test.valueType, Plain: true,
				}},
			}}
			m := New().SetBounds(0, 0, 60, 30)

			b.ResetTimer()
			for range b.N {
				m = m.SetData(data)
			}
		})
	}
}

func TestLargeValuePreviewsCropBeforeRendering(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	const tail = "tail must not be rendered"
	for _, test := range []struct {
		name      string
		value     string
		valueType string
	}{
		{name: "JSON", value: `{"payload":"` + strings.Repeat("abcdefghij", 100_000) + tail + `"}`, valueType: "JSON"},
		{name: "multiline string", value: strings.Repeat("large multiline value\n", 10_000) + tail, valueType: "STRING"},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := core.ParametersValue{
				Label: "default", Value: test.value, RawValue: test.value, ValueType: test.valueType, Plain: true,
			}
			lines := New().renderValueLines(value, 40)
			plain := ansi.Strip(strings.Join(lines, "\n"))
			if len(lines) > valuePreviewLineBudget {
				t.Fatalf("preview has %d lines, want at most %d", len(lines), valuePreviewLineBudget)
			}
			if strings.Contains(plain, tail) {
				t.Fatalf("preview contains cropped tail: %q", plain)
			}
			if !strings.HasSuffix(plain, "…") {
				t.Fatalf("preview has no truncation marker: %q", plain)
			}
			if len(plain) > 500 {
				t.Fatalf("preview retained excessive content: %d bytes", len(plain))
			}

			data := &messages.ParameterViewData{
				Parameter:        core.ParametersEntry{Key: "huge", Values: []core.ParametersValue{value}},
				SelectedValueIdx: 0,
			}
			model := New().SetBounds(0, 0, 60, 30).SetData(data)
			raw, ok := model.SelectedRawValue()
			if !ok || raw != test.value {
				t.Fatalf("editor raw value was cropped: ok=%v len=%d, want %d", ok, len(raw), len(test.value))
			}
		})
	}
}
