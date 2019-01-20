package domain

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"

	"github.com/srvc/fail"
)

// Preferences holds user specific preferences like timezone and privacy
type Preferences struct{}

// Value implements the driver.Valuer interface
func (p Preferences) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements the sql.Scanner interface
func (p *Preferences) Scan(src interface{}) error {
	value := reflect.ValueOf(src)
	if !value.IsValid() || value.IsNil() {
		return nil
	}

	if data, ok := src.([]byte); ok {
		var test []interface{}
		json.Unmarshal(data, &test)
		return json.Unmarshal(data, &p)
	}

	return fail.Errorf("could not not decode type %T -> %T", src, p)
}
