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
		return []byte("[]"), nil
	}
	return json.Marshal(j)
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

func (j JSONRaw) Value() (driver.Value, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	return []byte(j), nil
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
