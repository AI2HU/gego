package sqlutil

import (
	"database/sql"
	"encoding/json"
	"time"
)

func MapToJSON(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}
	data, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func JSONToMap(jsonStr string) map[string]string {
	if jsonStr == "" || jsonStr == "{}" {
		return make(map[string]string)
	}
	var result map[string]string
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return make(map[string]string)
	}
	if result == nil {
		return make(map[string]string)
	}
	return result
}

func SliceToJSON(slice []string) string {
	if len(slice) == 0 {
		return "[]"
	}
	data, err := json.Marshal(slice)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func JSONToSlice(jsonStr string) []string {
	if jsonStr == "" || jsonStr == "[]" {
		return []string{}
	}
	var result []string
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return []string{}
	}
	if result == nil {
		return []string{}
	}
	return result
}

func NullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}

func ScanNullableTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}
