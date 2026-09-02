package shared

import (
	"errors"
	"fmt"
	"io"
	"os"

	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/internal/terminal/progress"
	"github.com/yumauri/fbrcm/ops/invocation"
)

// ErrNoJSONSelection may be passed as onCancel to return fmt.Errorf("no %s selected", label).
var ErrNoJSONSelection = errors.New("no json selection")

// ReadJSONInput reads JSON bytes from fromPath, stdin, or an interactive file picker.
// When the picker is dismissed without a selection, onCancel controls the result:
// nil returns (nil, nil); ErrNoJSONSelection returns a label-based error; any other
// error is returned as-is.
func ReadJSONInput(cmd invocation.Call, fromPath, label string, onCancel error) ([]byte, error) {
	switch {
	case fromPath != "":
		data, err := os.ReadFile(fromPath)
		if err != nil {
			return nil, fmt.Errorf("read source file: %w", err)
		}
		return data, nil
	case StdinAvailable(cmd.InOrStdin()):
		data, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		return data, nil
	default:
		if MachineMode(cmd) {
			return nil, InteractionRequiredWithArguments(fmt.Sprintf("%s input requires --from or piped stdin in JSON mode", label), "external_input", false, "--from")
		}
		progress.Stop()
		selectedPath, err := PickFile([]string{".json"})
		if err != nil {
			return nil, err
		}
		if selectedPath == "" {
			switch {
			case onCancel == nil:
				return nil, nil
			case errors.Is(onCancel, ErrNoJSONSelection):
				return nil, fmt.Errorf("no %s selected", label)
			default:
				return nil, onCancel
			}
		}
		data, err := os.ReadFile(selectedPath)
		if err != nil {
			return nil, fmt.Errorf("read selected file: %w", err)
		}
		return data, nil
	}
}

type jsonPickerModel struct {
	picker   filepicker.Model
	selected string
	cancel   bool
}

// PickFile opens the shared interactive file selector, restricted to the
// supplied filename extensions. It returns an empty path when canceled.
func PickFile(allowedTypes []string) (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "."
	}

	picker := filepicker.New()
	picker.CurrentDirectory = currentDir
	picker.AllowedTypes = append([]string(nil), allowedTypes...)
	picker.FileAllowed = true
	picker.DirAllowed = false
	picker.ShowHidden = true
	picker.AutoHeight = true

	finalModel, err := tea.NewProgram(jsonPickerModel{picker: picker}).Run()
	if err != nil {
		return "", err
	}

	model, ok := finalModel.(jsonPickerModel)
	if !ok || model.cancel {
		return "", nil
	}
	return model.selected, nil
}

func (m jsonPickerModel) Init() tea.Cmd {
	return m.picker.Init()
}

func (m jsonPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancel = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.picker, cmd = m.picker.Update(msg)
	if didSelect, path := m.picker.DidSelectFile(msg); didSelect {
		m.selected = path
		return m, tea.Quit
	}

	return m, cmd
}

func (m jsonPickerModel) View() tea.View {
	return tea.NewView(m.picker.View() + "\n\nenter/l to open or select, h/backspace/left/esc to go up, q to cancel")
}
