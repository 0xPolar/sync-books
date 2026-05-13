package kavita

import (
	"fmt"
	"strings"
	"time"
)

// KavitaTime wraps time.Time with a tolerant JSON unmarshaller. Kavita is a
// .NET app and emits timestamps in two shapes the stdlib won't accept:
//   - "0001-01-01T00:00:00" (DateTime.MinValue, no zone) for unset fields
//   - "2024-01-02T15:04:05" (no zone) for some endpoints
//
// Unset values decode to the zero time.Time.
type KavitaTime struct {
	time.Time
}

var kavitaTimeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05.9999999",
	"2006-01-02T15:04:05",
}

func (t *KavitaTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" || strings.HasPrefix(s, "0001-01-01") {
		t.Time = time.Time{}
		return nil
	}
	for _, layout := range kavitaTimeLayouts {
		if v, err := time.Parse(layout, s); err == nil {
			t.Time = v
			return nil
		}
	}
	return fmt.Errorf("kavita: unrecognized time %q", s)
}
