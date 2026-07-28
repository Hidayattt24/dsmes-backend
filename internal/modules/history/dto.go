package history

// HistoryItemResponse represents a single activity/event in the patient's history timeline.
type HistoryItemResponse struct {
	ID           string         `json:"id"`
	PatientID    string         `json:"patient_id"`
	ActivityType string         `json:"activity_type"`
	Title        string         `json:"title"`
	Subtitle     string         `json:"subtitle"`
	Category     string         `json:"category"`
	Value        string         `json:"value"`
	Unit         string         `json:"unit"`
	Status       string         `json:"status"`
	Notes        string         `json:"notes"`
	MeasuredAt   string         `json:"measured_at"`
	CreatedAt    string         `json:"created_at"`
	UpdatedAt    string         `json:"updated_at"`
	RecordedBy   string         `json:"recorded_by,omitempty"`
	Icon         string         `json:"icon,omitempty"`
	Color        string         `json:"color,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}
