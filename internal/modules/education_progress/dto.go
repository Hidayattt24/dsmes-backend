package education_progress

// PatientProgressItem represents a single patient's progress for an education article.
type PatientProgressItem struct {
	PatientID        string  `json:"patient_id"`
	PatientName      string  `json:"patient_name"`
	Puskesmas        string  `json:"puskesmas"`
	ArticleRead      bool    `json:"article_read"`
	ArticleReadAt    *string `json:"article_read_at"`
	YouTubeWatched   bool    `json:"youtube_watched"`
	YouTubeWatchedAt *string `json:"youtube_watched_at"`
	Completed        bool    `json:"completed"`
	CompletedAt      *string `json:"completed_at"`
	LastActivityAt   *string `json:"last_activity_at"`
}

// ProgressAnalytics holds summary statistics for an education article.
type ProgressAnalytics struct {
	TotalPatients     int64 `json:"total_patients"`
	CompletedCount    int64 `json:"completed_count"`
	ReadArticleCount  int64 `json:"read_article_count"`
	WatchedVideoCount int64 `json:"watched_video_count"`
	ReadAndVideoCount int64 `json:"read_and_video_count"`
	NotStartedCount   int64 `json:"not_started_count"`
}

// MarkActionRequest is the request body for marking article read or video watched.
type MarkActionRequest struct {
	PatientID string `json:"patient_id" validate:"required,uuid4"`
}
