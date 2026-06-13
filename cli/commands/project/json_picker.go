package project

import (
	"os"

	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"
)

type pickerModel struct {
	picker   filepicker.Model
	selected string
	cancel   bool
}

func pickJSONFile() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "."
	}

	picker := filepicker.New()
	picker.CurrentDirectory = currentDir
	picker.AllowedTypes = []string{".json"}
	picker.FileAllowed = true
	picker.DirAllowed = false
	picker.ShowHidden = true
	picker.AutoHeight = true

	finalModel, err := tea.NewProgram(pickerModel{picker: picker}).Run()
	if err != nil {
		return "", err
	}

	model, ok := finalModel.(pickerModel)
	if !ok || model.cancel {
		return "", nil
	}
	return model.selected, nil
}

func (m pickerModel) Init() tea.Cmd {
	return m.picker.Init()
}

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m pickerModel) View() tea.View {
	return tea.NewView(m.picker.View() + "\n\nenter/l to open or select, h/backspace/left/esc to go up, q to cancel")
}
