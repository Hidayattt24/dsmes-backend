package education_progress

// PatientProgressItem represents a single patient's progress for an education article.
type PatientProgressItem struct {
	PatientID                 string  `json:"patient_id"`
	PatientName               string  `json:"patient_name"`
	Puskesmas                 string  `json:"puskesmas"`
	ArticleRead               bool    `json:"article_read"`
	ArticleReadAt             *string `json:"article_read_at"`
	ArticleStartedAt          *string `json:"article_started_at"`
	ArticleFinishedAt         *string `json:"article_finished_at"`
	ArticleReadingDuration    int     `json:"article_reading_duration"`
	ArticleLastScrollPosition int     `json:"article_last_scroll_position"`
	YouTubeWatched            bool    `json:"youtube_watched"`
	YouTubeWatchedAt          *string `json:"youtube_watched_at"`
	VideoStartedAt            *string `json:"video_started_at"`
	VideoFinishedAt           *string `json:"video_finished_at"`
	VideoWatchDuration        int     `json:"video_watch_duration"`
	VideoLastTimestamp        int     `json:"video_last_timestamp"`
	Completed                 bool    `json:"completed"`
	CompletedAt               *string `json:"completed_at"`
	LastActivityAt            *string `json:"last_activity_at"`
	CompletionSource          string  `json:"completion_source"`
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

type MarkArticleReadRequest struct {
	ReadingDuration    int `json:"reading_duration" validate:"gte=0"`
	LastScrollPosition int `json:"last_scroll_position" validate:"gte=0,lte=100"`
}

type MarkVideoWatchedRequest struct {
	WatchDuration      int `json:"watch_duration" validate:"gte=0"`
	VideoLastTimestamp int `json:"video_last_timestamp" validate:"gte=0"`
}
