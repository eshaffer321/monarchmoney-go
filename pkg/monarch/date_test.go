package monarch

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDate_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "date only format YYYY-MM-DD",
			input:   `"2025-08-30"`,
			want:    "2025-08-30",
			wantErr: false,
		},
		{
			name:    "RFC3339 format",
			input:   `"2025-08-30T15:04:05Z"`,
			want:    "2025-08-30",
			wantErr: false,
		},
		{
			name:    "datetime without timezone",
			input:   `"2025-08-30T15:04:05"`,
			want:    "2025-08-30",
			wantErr: false,
		},
		{
			name:    "null value",
			input:   `null`,
			want:    "",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   `""`,
			want:    "",
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   `"not-a-date"`,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Date
			err := json.Unmarshal([]byte(tt.input), &d)

			if (err != nil) != tt.wantErr {
				t.Errorf("Date.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				got := d.String()
				if got != tt.want {
					t.Errorf("Date.UnmarshalJSON() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDate_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		date Date
		want string
	}{
		{
			name: "normal date",
			date: Date{Time: time.Date(2025, 8, 30, 15, 30, 0, 0, time.UTC)},
			want: `"2025-08-30"`,
		},
		{
			name: "zero date",
			date: Date{Time: time.Time{}},
			want: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.date)
			if err != nil {
				t.Errorf("Date.MarshalJSON() error = %v", err)
				return
			}
			if string(got) != tt.want {
				t.Errorf("Date.MarshalJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestTransaction_DateParsing(t *testing.T) {
	// Test that Transaction struct properly unmarshals dates
	jsonData := `{
		"id": "123",
		"date": "2025-08-30",
		"amount": -50.00,
		"createdAt": "2025-08-30",
		"updatedAt": "2025-08-31",
		"reviewedAt": null
	}`

	var txn Transaction
	err := json.Unmarshal([]byte(jsonData), &txn)
	if err != nil {
		t.Fatalf("Failed to unmarshal transaction: %v", err)
	}

	// Check date field
	if txn.Date.String() != "2025-08-30" {
		t.Errorf("Transaction date = %v, want 2025-08-30", txn.Date.String())
	}

	// Check createdAt field
	if txn.CreatedAt.String() != "2025-08-30" {
		t.Errorf("Transaction createdAt = %v, want 2025-08-30", txn.CreatedAt.String())
	}

	// Check updatedAt field
	if txn.UpdatedAt.String() != "2025-08-31" {
		t.Errorf("Transaction updatedAt = %v, want 2025-08-31", txn.UpdatedAt.String())
	}

	// Check null reviewedAt field
	if txn.ReviewedAt != nil {
		t.Errorf("Transaction reviewedAt should be nil for null value")
	}
}
