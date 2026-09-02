package get

import (
	"testing"
	"time"
)

func TestPointerHelpers(t *testing.T) {
	if got := stringPtrOrNil(""); got != nil {
		t.Fatalf("stringPtrOrNil(blank) = %#v, want nil", got)
	}
	if got := stringPtrOrNil("value"); got == nil || *got != "value" {
		t.Fatalf("stringPtrOrNil(value) = %#v, want value pointer", got)
	}

	if got := timePtrOrNil(time.Time{}); got != nil {
		t.Fatalf("timePtrOrNil(zero) = %#v, want nil", got)
	}
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	if got := timePtrOrNil(now); got == nil || !got.Equal(now) {
		t.Fatalf("timePtrOrNil(now) = %#v, want now pointer", got)
	}
}
