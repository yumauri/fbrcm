package get

import "time"

func stringPtrOrNil(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}

func timePtrOrNil(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	v := value
	return &v
}
