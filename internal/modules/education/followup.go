package education

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/infrastructure/notifications"
)

const educationFollowupBatchSize = 100

type EducationFollowup struct {
	PatientID    string
	ArticleID    string
	ArticleTitle string
	OccurrenceID string
}

type EducationFollowupRepository interface {
	ClaimDue(ctx context.Context, now time.Time, limit int) ([]EducationFollowup, error)
	FindDeviceTokens(ctx context.Context, patientID string) ([]domain.DeviceToken, error)
	DeleteDeviceToken(ctx context.Context, patientID, token string) error
}

type educationFollowupRepository struct{ db *gorm.DB }

func NewEducationFollowupRepository(db *gorm.DB) EducationFollowupRepository {
	return &educationFollowupRepository{db: db}
}

// The INSERT is the claim: the unique occurrence index allows only one worker
// to create the inbox item and therefore to send the corresponding push.
func (r *educationFollowupRepository) ClaimDue(ctx context.Context, now time.Time, limit int) ([]EducationFollowup, error) {
	var items []EducationFollowup
	err := r.db.WithContext(ctx).Raw(`
WITH candidates AS (
    SELECT n.patient_id, n.article_id, a.title
    FROM notification_logs n
    JOIN articles a ON a.id = n.article_id
        AND a.deleted_at IS NULL AND a.status = 'publikasi'
    JOIN patients p ON p.id = n.patient_id AND p.deleted_at IS NULL
    WHERE n.education_followup_due_at <= ?
      AND n.notif_type = 'education'
      AND n.action_type = 'education_detail'
      AND n.occurrence_id = 'education:' || n.article_id::text || ':' || n.patient_id::text
      AND NOT EXISTS (
          SELECT 1 FROM article_views v
          WHERE v.article_id = n.article_id AND v.patient_id = n.patient_id
            AND v.deleted_at IS NULL
      )
      AND NOT EXISTS (
          SELECT 1 FROM notification_logs done
          WHERE done.occurrence_id = 'education-followup:' || n.article_id::text || ':' || n.patient_id::text
      )
    ORDER BY n.education_followup_due_at, n.id
    LIMIT ?
), inserted AS (
    INSERT INTO notification_logs (
        patient_id, article_id, message_text, notified_at,
        notif_type, action_type, occurrence_id, created_at, updated_at
    )
    SELECT c.patient_id, c.article_id,
        'Kamu belum membuka materi ini. Yuk pelajari saat sempat.', ?,
        'education', 'education_detail',
        'education-followup:' || c.article_id::text || ':' || c.patient_id::text,
        ?, ?
    FROM candidates c
    WHERE true
    ON CONFLICT (occurrence_id) WHERE occurrence_id IS NOT NULL DO NOTHING
    RETURNING patient_id, article_id, occurrence_id
)
SELECT i.patient_id, i.article_id, a.title AS article_title, i.occurrence_id
FROM inserted i JOIN articles a ON a.id = i.article_id
`, now, limit, now, now, now).Scan(&items).Error
	return items, err
}

func (r *educationFollowupRepository) FindDeviceTokens(ctx context.Context, patientID string) ([]domain.DeviceToken, error) {
	var devices []domain.DeviceToken
	err := r.db.WithContext(ctx).
		Table("patient_device_tokens AS pdt").
		Select("DISTINCT pdt.token, pdt.patient_id, pdt.platform, pdt.last_seen, COALESCE(NULLIF(TRIM(p.nickname), ''), p.full_name, '') AS notification_name").
		Joins("JOIN patients p ON p.id = pdt.patient_id AND p.deleted_at IS NULL").
		Where("pdt.patient_id = ?", patientID).
		Find(&devices).Error
	return devices, err
}

func (r *educationFollowupRepository) DeleteDeviceToken(ctx context.Context, patientID, token string) error {
	return r.db.WithContext(ctx).Where("patient_id = ? AND token = ?", patientID, token).Delete(&domain.DeviceToken{}).Error
}

func DispatchDueEducationFollowups(ctx context.Context, repo EducationFollowupRepository, sender PushSender, now time.Time, logger *zap.Logger) error {
	items, err := repo.ClaimDue(ctx, now.UTC(), educationFollowupBatchSize)
	if err != nil {
		return err
	}
	var targets, sent, failed atomic.Int64
	semaphore := make(chan struct{}, 4)
	var group sync.WaitGroup
	for _, item := range items {
		devices, err := repo.FindDeviceTokens(ctx, item.PatientID)
		if err != nil {
			logger.Warn("education followup device lookup failed", zap.String("article_id", item.ArticleID), zap.String("patient_id", item.PatientID), zap.Error(err))
			continue
		}
		seen := make(map[string]bool, len(devices))
		for _, device := range devices {
			if seen[device.Token] {
				continue
			}
			seen[device.Token] = true
			targets.Add(1)
			group.Add(1)
			go func(item EducationFollowup, device domain.DeviceToken) {
				defer group.Done()
				select {
				case semaphore <- struct{}{}:
					defer func() { <-semaphore }()
				case <-ctx.Done():
					failed.Add(1)
					return
				}
				body := "Kamu belum membuka materi ini. Yuk pelajari saat sempat."
				if name := strings.TrimSpace(device.NotificationName); name != "" {
					body = fmt.Sprintf("Hai, %s! %s", name, body)
				}
				deviceCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
				defer cancel()
				_, err := sender.Send(deviceCtx, device.Token,
					fmt.Sprintf("Pengingat Edukasi: %s 📖", item.ArticleTitle), body,
					map[string]string{
						"type": "education", "action_type": "education_detail",
						"article_id": item.ArticleID, "occurrence_id": item.OccurrenceID,
						"sent_at": time.Now().UTC().Format(time.RFC3339Nano),
					})
				if err != nil {
					failed.Add(1)
					logger.Warn("education followup push failed", zap.String("article_id", item.ArticleID), zap.String("patient_id", item.PatientID), zap.Error(err))
					if notifications.IsUnregistered(err) || strings.EqualFold(err.Error(), "NotRegistered") {
						cleanupCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
						defer stop()
						if deleteErr := repo.DeleteDeviceToken(cleanupCtx, item.PatientID, device.Token); deleteErr != nil {
							logger.Warn("education followup invalid token cleanup failed", zap.String("patient_id", item.PatientID), zap.Error(deleteErr))
						}
					}
					return
				}
				sent.Add(1)
			}(item, device)
		}
	}
	group.Wait()
	if len(items) > 0 {
		logger.Info("education followup dispatch completed", zap.Int("inbox_created", len(items)), zap.Int64("device_targets", targets.Load()), zap.Int64("push_sent", sent.Load()), zap.Int64("push_failed", failed.Load()))
	}
	return nil
}
