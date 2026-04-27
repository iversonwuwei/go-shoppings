package model

import (
	"database/sql/driver"
	"encoding/json"
	"testing"
)

func TestJSONBValueReturnsJSONString(t *testing.T) {
	value, err := JSONB([]string{"express", "self_pickup"}).Value()
	if err != nil {
		t.Fatalf("Value returned error: %v", err)
	}
	if _, ok := value.(string); !ok {
		t.Fatalf("Value type = %T, want string", value)
	}
	if driver.Value(value) != `["express","self_pickup"]` {
		t.Fatalf("Value = %v", value)
	}
}

func TestJSONRawJSONRoundTrip(t *testing.T) {
	var raw JSONRaw
	if err := json.Unmarshal([]byte(`{"color":"red","size":"L"}`), &raw); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	value, err := raw.Value()
	if err != nil {
		t.Fatalf("Value returned error: %v", err)
	}
	if _, ok := value.(string); !ok {
		t.Fatalf("Value type = %T, want string", value)
	}
	if value != `{"color":"red","size":"L"}` {
		t.Fatalf("Value = %v", value)
	}

	encoded, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if string(encoded) != `{"color":"red","size":"L"}` {
		t.Fatalf("Marshal = %s", encoded)
	}
}
