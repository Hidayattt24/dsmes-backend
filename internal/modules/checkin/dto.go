package checkin

type CheckinRequest struct {
	CheckinDate string `json:"checkin_date" validate:"required"` // format: "YYYY-MM-DD"
}

type CheckinResponse struct {
	ID          string `json:"id"`
	CheckinDate string `json:"checkin_date"`
	IsCompleted bool   `json:"is_completed"`
}

type CheckinCalendarResponse struct {
	CompletedDates []string `json:"completed_dates"`
}
