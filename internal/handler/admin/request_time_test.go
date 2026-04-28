package admin

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseRequestTimeAcceptsLocalDateTime(t *testing.T) {
	parsed, err := parseRequestTime("2026-04-27T00:00:00")
	if err != nil {
		t.Fatalf("parseRequestTime returned error: %v", err)
	}
	if parsed.Year() != 2026 || parsed.Month() != time.April || parsed.Day() != 27 || parsed.Hour() != 0 {
		t.Fatalf("unexpected parsed time: %s", parsed.Format(time.RFC3339))
	}
}

func TestParseRequestTimeAcceptsRFC3339(t *testing.T) {
	parsed, err := parseRequestTime("2026-04-27T00:00:00+08:00")
	if err != nil {
		t.Fatalf("parseRequestTime returned error: %v", err)
	}
	_, offset := parsed.Zone()
	if offset != 8*60*60 {
		t.Fatalf("unexpected timezone offset: %d", offset)
	}
}

func TestRequestTimeUnmarshalNullAndInvalid(t *testing.T) {
	var empty *requestTime
	if err := json.Unmarshal([]byte(`null`), &empty); err != nil {
		t.Fatalf("null should unmarshal without error: %v", err)
	}
	if empty != nil {
		t.Fatal("null should keep requestTime pointer nil")
	}

	var invalid requestTime
	if err := json.Unmarshal([]byte(`"not-a-date"`), &invalid); err == nil {
		t.Fatal("invalid datetime should return error")
	}
}