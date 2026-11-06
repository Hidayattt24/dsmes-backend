package education

import (
	"os"
	"testing"

	"github.com/dsmes/dsmes-backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Run explicitly against the configured database to inspect the read-only plan.
func TestEducationListQueryPlan(t *testing.T) {
	if os.Getenv("RUN_EDUCATION_EXPLAIN") != "1" {
		t.Skip("set RUN_EDUCATION_EXPLAIN=1 to inspect database plan")
	}
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(previousDir)
	if err := os.Chdir("../../.."); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(postgres.Open(cfg.DB.DSN()), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	rows, err := sqlDB.Query(`EXPLAIN (ANALYZE, BUFFERS)
		SELECT articles.id, articles.title, articles.category_id,
		articles.estimated_read_minutes, articles.estimated_read_seconds,
		articles.author_name, articles.banner_image_url, articles.summary,
		articles.status, articles.created_at, articles.updated_at,
		articles.youtube_link, articles.audio_url,
		(SELECT count(*) FROM article_views v WHERE v.article_id = articles.id AND v.deleted_at IS NULL) AS read_count
		FROM articles WHERE status = 'publikasi' AND deleted_at IS NULL
		ORDER BY created_at DESC LIMIT 10`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatal(err)
		}
		t.Log(line)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	var sizes []struct {
		ID          string
		Content     int64
		Banner      int64
		BannerKind  string
	}
	if err := db.Raw(`SELECT id, octet_length(content) AS content,
			octet_length(banner_image_url) AS banner,
			CASE WHEN banner_image_url LIKE 'data:image/%' THEN 'data-uri'
			ELSE 'url' END AS banner_kind
			FROM articles WHERE status = 'publikasi' AND deleted_at IS NULL
			ORDER BY created_at DESC LIMIT 10`).Scan(&sizes).Error; err != nil {
		t.Fatal(err)
	}
	for _, size := range sizes {
		t.Logf("article=%s content_bytes=%d banner_bytes=%d banner_kind=%s", size.ID, size.Content, size.Banner, size.BannerKind)
	}
}
