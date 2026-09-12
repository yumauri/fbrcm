package setup

import (
	"context"
	"errors"
	"fmt"

	"github.com/erikgeiser/promptkit"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/ops/shared"
)

type promptChoice struct {
	Label string
	Value string
}

func (c promptChoice) String() string { return c.Label }

type prompter interface {
	selectOne(prompt string, choices []promptChoice) (string, error)
	text(prompt, initial string, validate func(string) error) (string, error)
	pickJSON(label string) (string, error)
}

type terminalPrompter struct{ cmd *cobra.Command }

func newTerminalPrompter(cmd *cobra.Command) prompter { return &terminalPrompter{cmd: cmd} }

func (p *terminalPrompter) selectOne(prompt string, choices []promptChoice) (string, error) {
	input, closeInput, err := shared.OpenPromptInput(p.cmd.InOrStdin())
	if err != nil {
		return "", err
	}
	defer closeInput()

	menu := selection.New(prompt, choices)
	menu.Filter = nil
	menu.Input = input
	menu.Output = p.cmd.ErrOrStderr()
	choice, err := menu.RunPrompt()
	if err != nil {
		return "", normalizePromptError(err)
	}
	return choice.Value, nil
}

func (p *terminalPrompter) text(prompt, initial string, validate func(string) error) (string, error) {
	input, closeInput, err := shared.OpenPromptInput(p.cmd.InOrStdin())
	if err != nil {
		return "", err
	}
	defer closeInput()

	field := textinput.New(prompt)
	field.InitialValue = initial
	field.Validate = validate
	field.Input = input
	field.Output = p.cmd.ErrOrStderr()
	value, err := field.RunPrompt()
	if err != nil {
		return "", normalizePromptError(err)
	}
	return value, nil
}

func (p *terminalPrompter) pickJSON(label string) (string, error) {
	_, _ = fmt.Fprintf(p.cmd.ErrOrStderr(), "Select %s.\n", label)
	path, err := shared.PickFile([]string{".json"})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", context.Canceled
	}
	return path, nil
}

func normalizePromptError(err error) error {
	if errors.Is(err, promptkit.ErrAborted) {
		return context.Canceled
	}
	return err
}
