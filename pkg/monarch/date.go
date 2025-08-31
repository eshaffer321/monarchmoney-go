package monarch

import (
	"fmt"
	"strings"
	"time"
)

// Date is a custom type that handles date-only JSON values
type Date struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler for Date
func (d *Date) UnmarshalJSON(data []byte) error {
	// Remove quotes
	str := strings.Trim(string(data), `"`)

	// Handle null/empty
	if str == "" || str == "null" {
		d.Time = time.Time{}
		return nil
	}

	// Try parsing as date only first (YYYY-MM-DD)
	t, err := time.Parse("2006-01-02", str)
	if err == nil {
		d.Time = t
		return nil
	}

	// Try parsing as full timestamp (RFC3339)
	t, err = time.Parse(time.RFC3339, str)
	if err == nil {
		d.Time = t
		return nil
	}

	// Try parsing with time but no timezone
	t, err = time.Parse("2006-01-02T15:04:05", str)
	if err == nil {
		d.Time = t
		return nil
	}

	return fmt.Errorf("unable to parse date: %s", str)
}

// MarshalJSON implements json.Marshaler for Date
func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	// Format as date only
	return []byte(fmt.Sprintf(`"%s"`, d.Time.Format("2006-01-02"))), nil
}

// String returns the date as a string
func (d Date) String() string {
	if d.Time.IsZero() {
		return ""
	}
	return d.Time.Format("2006-01-02")
}
