package explain

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/erikgeiser/promptkit"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"

	"github.com/yumauri/fbrcm/core/rc/evaluate"
	"github.com/yumauri/fbrcm/internal/terminal/progress"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
)

type terminalPrompter struct{ cmd invocation.Call }

func newPrompter(cmd invocation.Call) prompter { return terminalPrompter{cmd: cmd} }

func (p terminalPrompter) input() (io.Reader, func(), error) {
	progress.Stop()
	if err := shared.CommandContext(p.cmd).Err(); err != nil {
		return nil, func() {}, err
	}
	return shared.OpenPromptInput(p.cmd.InOrStdin())
}

func (p terminalPrompter) Fact(need evaluate.Need) (evaluate.Fact, error) {
	input, closeInput, err := p.input()
	if err != nil {
		return evaluate.Fact{}, fmt.Errorf("cannot ask for %s: %w; use --set or --no-input", need.Key, err)
	}
	defer closeInput()
	_, _ = fmt.Fprintf(p.cmd.ErrOrStderr(), "\nMissing input: %s\nEnter a value, /unknown, or /unset.\n", need.Label)
	if need.Kind == "boolean" {
		_, _ = fmt.Fprintln(p.cmd.ErrOrStderr(), "Use true or false for membership.")
	}
	field := textinput.New(need.Key + ":")
	field.Validate = func(value string) error {
		_, err := evaluate.Answer(need, value, "your answer")
		return err
	}
	field.Input, field.Output = input, p.cmd.ErrOrStderr()
	answer, err := field.RunPrompt()
	if err != nil {
		return evaluate.Fact{}, promptError(err)
	}
	return evaluate.Answer(need, answer, "your answer")
}

type exchangeChoice struct {
	label string
	value evaluate.Exchange
}

func (c exchangeChoice) String() string { return c.label }

func (p terminalPrompter) Exchange(exchanges []evaluate.Exchange) (evaluate.Exchange, error) {
	input, closeInput, err := p.input()
	if err != nil {
		return evaluate.Exchange{}, fmt.Errorf("capture has multiple Firebase fetches; choose an original HAR entry with --entry")
	}
	defer closeInput()
	choices := make([]exchangeChoice, len(exchanges))
	for i, exchange := range exchanges {
		label := fmt.Sprintf("[%d] %s · %s · app %s · HTTP %d", exchange.Entry, exchange.Project, exchange.StartedAt.Format("15:04:05"), exchange.Facts["app.id"].Value, exchange.Status)
		choices[i] = exchangeChoice{label: label, value: exchange}
	}
	menu := selection.New("Choose a Firebase fetch", choices)
	menu.Filter = nil
	menu.Input, menu.Output = input, p.cmd.ErrOrStderr()
	choice, err := menu.RunPrompt()
	return choice.value, promptError(err)
}

func promptError(err error) error {
	if errors.Is(err, promptkit.ErrAborted) {
		return context.Canceled
	}
	return err
}
