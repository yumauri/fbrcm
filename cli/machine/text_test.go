package machine

import (
	"errors"
	"strings"
	"testing"
)

func TestSafeTextRedactsAndBoundsMachineText(t *testing.T) {
	secret := "super-secret-value"
	long := strings.Repeat("x", MaxSafeTextRunes+100)
	got := SafeErrorText(errors.New(`failed: {"access_token":"` + secret + `","client_id":"` + secret + `"} bearer ` + secret + ` token=` + secret + ` client_id=` + secret + ` ` + long))
	if strings.Contains(got, secret) || !strings.Contains(got, "[REDACTED]") {
		t.Fatalf("SafeErrorText exposed a secret: %q", got)
	}
	if len([]rune(got)) != MaxSafeTextRunes+1 || !strings.HasSuffix(got, "…") {
		t.Fatalf("SafeErrorText length = %d, value = %q", len([]rune(got)), got)
	}
}
