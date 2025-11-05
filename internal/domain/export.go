package domain

import "time"

// ExportBundle represents an export of a profile's data
type ExportBundle struct {
	ID        string    `json:"id"`
	ProfileID string    `json:"profile_id"`
	Format    string    `json:"format"`   // json, zip, csv
	Location  string    `json:"location"` // Storage URL or path
	CreatedAt time.Time `json:"created_at"`
}

// IsJSON checks if the export is in JSON format
func (e *ExportBundle) IsJSON() bool {
	return e.Format == ExportFormatJSON
}

// IsZIP checks if the export is in ZIP format
func (e *ExportBundle) IsZIP() bool {
	return e.Format == ExportFormatZIP
}

// IsCSV checks if the export is in CSV format
func (e *ExportBundle) IsCSV() bool {
	return e.Format == ExportFormatCSV
}

// Export format constants
const (
	ExportFormatJSON = "json"
	ExportFormatZIP  = "zip"
	ExportFormatCSV  = "csv"
)
