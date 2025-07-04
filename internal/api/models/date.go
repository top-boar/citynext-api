package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// Date represents a date (YYYY-MM-DD) with no time component
// Implements JSON and SQL interfaces
// Always stores time as UTC midnight

type Date struct {
	time.Time
}

const dateLayout = "2006-01-02"

// parses a date string (YYYY-MM-DD) into a Date object
func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		d.Time = time.Time{}
		return nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("invalid date: %w", err)
	}
	d.Time = t.UTC()
	return nil
}

// formats the date as YYYY-MM-DD
func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte(`null`), nil
	}
	return []byte(`"` + d.Time.Format(dateLayout) + `"`), nil
}

// implements the driver.Value interface for database serialization
func (d Date) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time.Format(dateLayout), nil
}

// implements the sql.Scanner interface for database deserialization
func (d *Date) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		d.Time = v.UTC()
		return nil
	case string:
		t, err := time.Parse(dateLayout, v)
		if err != nil {
			return err
		}
		d.Time = t.UTC()
		return nil
	case []byte:
		t, err := time.Parse(dateLayout, string(v))
		if err != nil {
			return err
		}
		d.Time = t.UTC()
		return nil
	}
	return fmt.Errorf("cannot scan type %T into Date", value)
}

// returns the date as YYYY-MM-DD
func (d Date) String() string {
	return d.Time.Format(dateLayout)
}
