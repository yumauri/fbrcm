package buttonbar

type Variant int

const (
	VariantNeutral Variant = iota
	VariantDanger
	VariantAccent
)

type Button struct {
	Label   string
	Variant Variant
}

type Model struct {
	buttons  []Button
	selected int
	focused  bool
}

func New(buttons []Button) Model {
	return Model{buttons: append([]Button(nil), buttons...)}
}

func (m Model) SetSelected(selected int) Model {
	m.selected = min(max(selected, 0), max(len(m.buttons)-1, 0))
	return m
}

func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
}

func (m Model) Selected() int { return m.selected }

func (m *Model) Move(delta int) {
	if len(m.buttons) == 0 {
		return
	}
	m.selected = (m.selected + delta + len(m.buttons)) % len(m.buttons)
}

func (m Model) IndexAt(x, y int) (int, bool) {
	buttonX := 0
	for i, button := range m.rendered() {
		width, height := printableWidth(button), renderedHeight(button)
		if x >= buttonX && x < buttonX+width && y >= 0 && y < height {
			return i, true
		}
		buttonX += width
		if i < len(m.buttons)-1 {
			buttonX++
		}
	}
	return -1, false
}
