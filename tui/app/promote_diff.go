package app

import (
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/dictdiff"
	rcpromote "github.com/yumauri/fbrcm/core/rc/promote"
)

func (m *Model) openPromotionDiff(item rcpromote.Item) {
	input, ok := m.promote.DiffInput(item)
	if !ok {
		m.openErrorDialog("Diff Unavailable", m.promote.Source(), "The selected promotion change cannot be prepared for comparison.")
		return
	}
	m.openDictionaryDiff(input, m.promote.Source())
}

func (m *Model) openDictionaryDiff(input dictdiff.Input, project core.Project) {
	result, err := dictdiff.Compare(input)
	if err != nil {
		m.openErrorDialog("Diff Unavailable", project, err.Error())
		return
	}
	m.diffView = m.diffView.Open(m.width, m.height, result)
}
