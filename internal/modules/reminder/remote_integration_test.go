package reminder

import (
	"context"
	"os"
	"sort"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/bootstrap"
	"github.com/dsmes/dsmes-backend/internal/config"
	"github.com/dsmes/dsmes-backend/internal/domain"
)

// TestRemoteReminderFlow exercises the migrated reminder schema against the
// configured database. All writes happen inside a transaction that is always
// rolled back. Enable explicitly with RUN_REMOTE_REMINDER_TEST=1.
func TestRemoteReminderFlow(t *testing.T) {
	if os.Getenv("RUN_REMOTE_REMINDER_TEST") != "1" {
		t.Skip("set RUN_REMOTE_REMINDER_TEST=1 to run the remote rollback test")
	}
	// `go test` runs with the package directory as CWD; move to the backend
	// root so config.Load can discover the repository's .env file.
	t.Chdir("../../..")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	logger := zap.NewNop()
	db, err := bootstrap.NewDatabase(cfg, logger)
	if err != nil {
		t.Fatal(err)
	}
	defer bootstrap.CloseDatabase(db, logger)

	var patientID string
	if err := db.Raw("SELECT id FROM patients WHERE deleted_at IS NULL ORDER BY created_at LIMIT 1").Scan(&patientID).Error; err != nil {
		t.Fatal(err)
	}
	if patientID == "" {
		t.Skip("remote database has no active patient")
	}
	var initialSystemCount int64
	if err := db.Model(&domain.Reminder{}).Where("patient_id = ? AND system_key IS NOT NULL", patientID).Count(&initialSystemCount).Error; err != nil {
		t.Fatal(err)
	}
	var occurrenceID string

	errRollback := db.Transaction(func(tx *gorm.DB) error {
		repo := NewReminderRepository(tx, logger)
		ctx := context.Background()

		if err := repo.EnsureDefaultReminders(ctx, patientID); err != nil {
			return err
		}
		if err := repo.EnsureDefaultReminders(ctx, patientID); err != nil {
			return err
		}

		var reminders []domain.Reminder
		if err := tx.Preload("ActiveDays").Where("patient_id = ? AND system_key IS NOT NULL", patientID).Find(&reminders).Error; err != nil {
			return err
		}
		byKey := make(map[string]domain.Reminder)
		for _, item := range reminders {
			if item.SystemKey != nil {
				byKey[*item.SystemKey] = item
			}
		}
		assertReminder := func(key, at, action string, days []int) {
			t.Helper()
			item, ok := byKey[key]
			if !ok {
				t.Fatalf("missing system reminder %s", key)
			}
			if item.ScheduledTime != at || item.ActionType != action || !item.IsActive {
				t.Fatalf("%s got time=%s action=%s active=%v", key, item.ScheduledTime, item.ActionType, item.IsActive)
			}
			gotDays := make([]int, 0, len(item.ActiveDays))
			for _, day := range item.ActiveDays {
				gotDays = append(gotDays, day.DayOfWeek)
			}
			sort.Ints(gotDays)
			if len(gotDays) != len(days) {
				t.Fatalf("%s got days=%v want=%v", key, gotDays, days)
			}
			for i := range days {
				if gotDays[i] != days[i] {
					t.Fatalf("%s got days=%v want=%v", key, gotDays, days)
				}
			}
		}
		assertReminder("daily_meal", "12:00:00", "meal_entry", []int{1, 2, 3, 4, 5, 6, 7})
		assertReminder("daily_exercise", "16:00:00", "activity_entry", []int{1, 2, 3, 4, 5, 6, 7})
		assertReminder("weekly_blood_sugar", "07:00:00", "blood_sugar_entry", []int{6})

		meal := byKey["daily_meal"]
		jakarta, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			return err
		}
		nowJakarta := time.Now().In(jakarta)
		meal.ScheduledTime = nowJakarta.Format("15:04:00")
		if err := tx.Model(&domain.Reminder{}).Where("id = ?", meal.ID).Update("scheduled_time", meal.ScheduledTime).Error; err != nil {
			return err
		}
		due, err := repo.FindDueReminders(ctx, nowJakarta.Format("15:04"))
		if err != nil {
			return err
		}
		foundDue := false
		for _, item := range due {
			if item.ID == meal.ID && len(item.ActiveDays) == 7 {
				foundDue = true
				break
			}
		}
		if !foundDue {
			t.Fatalf("worker due query did not return test meal reminder at %s", nowJakarta.Format("15:04"))
		}

		scheduled := time.Now().UTC().Truncate(time.Minute)
		occurrenceID = "integration-test:" + meal.ID + ":" + scheduled.Format(time.RFC3339)
		log := &domain.NotificationLog{
			ReminderID: &meal.ID, PatientID: patientID, MessageText: "integration test",
			NotifiedAt: scheduled, NotifType: "reminder", ActionType: meal.ActionType,
			OccurrenceID: &occurrenceID, ScheduledAt: &scheduled,
		}
		created, err := repo.CreateNotificationOccurrence(ctx, log)
		if err != nil || !created {
			t.Fatalf("first occurrence: created=%v err=%v", created, err)
		}
		duplicate := *log
		duplicate.ID = ""
		created, err = repo.CreateNotificationOccurrence(ctx, &duplicate)
		if err != nil || created {
			t.Fatalf("duplicate occurrence: created=%v err=%v", created, err)
		}

		return errIntentionalRollback
	})
	if errRollback != errIntentionalRollback {
		t.Fatalf("remote rollback test failed: %v", errRollback)
	}
	var finalSystemCount, finalOccurrenceCount int64
	if err := db.Model(&domain.Reminder{}).Where("patient_id = ? AND system_key IS NOT NULL", patientID).Count(&finalSystemCount).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&domain.NotificationLog{}).Where("occurrence_id = ?", occurrenceID).Count(&finalOccurrenceCount).Error; err != nil {
		t.Fatal(err)
	}
	if finalSystemCount != initialSystemCount || finalOccurrenceCount != 0 {
		t.Fatalf("rollback leaked data: reminders before=%d after=%d occurrences=%d", initialSystemCount, finalSystemCount, finalOccurrenceCount)
	}
}

var errIntentionalRollback = &rollbackError{}

type rollbackError struct{}

func (*rollbackError) Error() string { return "intentional integration-test rollback" }
