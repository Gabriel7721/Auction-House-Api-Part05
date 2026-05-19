package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}

	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	return string(data), nil
}

func (s *StringArray) Scan(value any) error {
	if value == nil {
		*s = StringArray{}
		return nil
	}

	var bytes []byte

	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into StringArray", value)
	}

	if len(bytes) == 0 {
		*s = StringArray{}
		return nil
	}

	return json.Unmarshal(bytes, s)
}
