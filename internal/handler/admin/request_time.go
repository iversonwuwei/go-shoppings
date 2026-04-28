package admin

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type requestTime struct {
	time.Time
}

func (t *requestTime) UnmarshalJSON(data []byte) error {
	var raw *string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil
	}
	parsed, err := parseRequestTime(*raw)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

func parseRequestTime(value string) (time.Time, error) {
	s := strings.TrimSpace(value)
	if s == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.Parse(layout, s); err == nil {
			return parsed, nil
		}
	}
	for _, layout := range []string{"2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02"} {
		if parsed, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid datetime %q, expected RFC3339 or YYYY-MM-DDTHH:mm:ss", value)
}

func requestTimePtr(value *requestTime) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	parsed := value.Time
	return &parsed
}
