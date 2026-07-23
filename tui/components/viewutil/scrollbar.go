package viewutil

type Scrollbar struct {
	Visible    bool
	ThumbStart int
	ThumbEnd   int
}

// ScrollbarState maps a content offset to a proportional scrollbar thumb.
func ScrollbarState(total, offset, visible int) Scrollbar {
	if visible <= 0 || total <= visible {
		return Scrollbar{}
	}
	thumbHeight := max(1, (visible*visible)/total)
	maxThumbStart := visible - thumbHeight
	maxOffset := max(total-visible, 1)
	thumbStart := (min(offset, maxOffset) * maxThumbStart) / maxOffset
	return Scrollbar{
		Visible:    true,
		ThumbStart: thumbStart,
		ThumbEnd:   thumbStart + thumbHeight - 1,
	}
}
