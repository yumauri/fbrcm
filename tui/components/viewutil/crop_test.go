package viewutil

import (
	"strings"
	"testing"
)

func TestCropTextBeforeRenderStopsAtDisplayBudget(t *testing.T) {
	value := "ab界\nefghijkl"
	fragment, cropped := CropTextBeforeRender(value, 4, 3, 2)
	if !cropped || fragment != "ab界\nefg" {
		t.Fatalf("CropTextBeforeRender() = %q, %v; want %q, true", fragment, cropped, "ab界\nefg")
	}
}

func TestCropTextBeforeRenderDoesNotScanHugeTail(t *testing.T) {
	value := strings.Repeat("abcdefghij", 100_000) + "tail must not be rendered"
	fragment, cropped := CropTextBeforeRender(value, 12, 16, 5)
	if !cropped {
		t.Fatal("huge value was not cropped")
	}
	if len(fragment) > 100 || strings.Contains(fragment, "tail must not be rendered") {
		t.Fatalf("fragment contains too much input: len=%d tail=%v", len(fragment), strings.Contains(fragment, "tail must not be rendered"))
	}
}
