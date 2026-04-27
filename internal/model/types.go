package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONB 用于映射 PostgreSQL jsonb 列 => Go []string / map
// 这里只支持 []string（满足 features / images / delivery_type），其余字段可用 JSONRaw
type JSONB []string

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	bs, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(bs), nil
}

func (j *JSONB) Scan(v interface{}) error {
	if v == nil {
		*j = nil
		return nil
	}
	var bs []byte
	switch t := v.(type) {
	case []byte:
		bs = t
	case string:
		bs = []byte(t)
	default:
		return errors.New("invalid jsonb source")
	}
	if len(bs) == 0 {
		*j = nil
		return nil
	}
	return json.Unmarshal(bs, j)
}

// JSONRaw 任意 jsonb
type JSONRaw []byte

func (j JSONRaw) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	if !json.Valid(j) {
		return nil, errors.New("invalid json raw value")
	}
	return j, nil
}

func (j *JSONRaw) UnmarshalJSON(v []byte) error {
	if len(v) == 0 || string(v) == "null" {
		*j = nil
		return nil
	}
	if !json.Valid(v) {
		return errors.New("invalid json raw value")
	}
	*j = append((*j)[0:0], v...)
	return nil
}

func (j JSONRaw) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "{}", nil
	}
	if !json.Valid(j) {
		return nil, errors.New("invalid json raw value")
	}
	return string(j), nil
}

func (j *JSONRaw) Scan(v interface{}) error {
	if v == nil {
		*j = nil
		return nil
	}
	switch t := v.(type) {
	case []byte:
		*j = append((*j)[0:0], t...)
	case string:
		*j = []byte(t)
	default:
		return errors.New("invalid jsonb source")
	}
	return nil
}
