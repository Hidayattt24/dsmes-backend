package main

import (
	"context"
	"testing"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/modules/reminder"
	"go.uber.org/zap"
)

type dispatchTestRepo struct {
	reminder.ReminderRepository
	item    domain.Reminder
	seen    map[string]bool
	created int
}

func (r *dispatchTestRepo) FindDueReminders(_ context.Context, minute string) ([]domain.Reminder, error) {
	if r.item.ScheduledTime[:5] == minute {
		return []domain.Reminder{r.item}, nil
	}
	return nil, nil
}
func (r *dispatchTestRepo) CreateNotificationOccurrence(_ context.Context, log *domain.NotificationLog) (bool, error) {
	if r.seen[*log.OccurrenceID] {
		return false, nil
	}
	r.seen[*log.OccurrenceID] = true
	r.created++
	return true, nil
}
func (r *dispatchTestRepo) FindDeviceTokens(context.Context, string) ([]domain.DeviceToken, error) {
	return []domain.DeviceToken{{Token: "test-token", PatientID: "patient"}}, nil
}

type dispatchTestSender struct {
	sent     int
	lastData map[string]string
}

func (s *dispatchTestSender) Send(_ context.Context, _, _, _ string, data map[string]string) (string, error) {
	s.sent++
	s.lastData = data
	return "message-id", nil
}

func TestDispatchSchedulesAndDeduplicates(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		key, date, clock string
		day              int
	}{
		{"daily_meal", "2026-11-03", "12:00", 2},
		{"daily_exercise", "2026-11-03", "16:00", 2},
		{"weekly_blood_sugar", "2026-11-07", "07:00", 6},
	} {
		t.Run(tc.key, func(t *testing.T) {
			at, err := time.ParseInLocation("2006-01-02 15:04", tc.date+" "+tc.clock, loc)
			if err != nil {
				t.Fatal(err)
			}
			key := tc.key
			repo := &dispatchTestRepo{item: domain.Reminder{
				BaseModel: domain.BaseModel{ID: "reminder"}, PatientID: "patient", SystemKey: &key,
				ScheduledTime: tc.clock + ":00", IsActive: true, ActionType: "blood_sugar_entry",
				ActiveDays: []domain.ReminderActiveDay{{DayOfWeek: tc.day}},
			}, seen: map[string]bool{}}
			sender := &dispatchTestSender{}
			for _, minute := range catchUpMinutes(at.Add(3*time.Minute), 5) {
				if err := dispatchDueReminders(context.Background(), repo, sender, minute, zap.NewNop()); err != nil {
					t.Fatal(err)
				}
			}
			// A second scan of the same occurrence must not produce another inbox item or push.
			if err := dispatchDueReminders(context.Background(), repo, sender, at, zap.NewNop()); err != nil {
				t.Fatal(err)
			}
			if repo.created != 1 || sender.sent != 1 {
				t.Fatalf("inbox=%d push=%d, want 1 each", repo.created, sender.sent)
			}
			if sender.lastData["scheduled_at"] == "" {
				t.Fatal("expected scheduled_at in payload data")
			}
			if sender.lastData["sent_at"] == "" {
				t.Fatal("expected sent_at in payload data")
			}
			if sender.lastData["occurrence_id"] == "" {
				t.Fatal("expected occurrence_id in payload data")
			}
		})
	}
}

func TestBuildSystemReminderNotification(t *testing.T) {
	tests := []struct {
		key, wantTitle, wantBody string
	}{
		{"daily_meal", "Waktunya Catat Makanan 🥗", "Jangan lupa mencatat makanan yang kamu konsumsi hari ini, ya."},
		{"daily_exercise", "Waktunya Bergerak 🏃", "Yuk luangkan waktu untuk berolahraga dan tetap aktif sore ini."},
		{"weekly_blood_sugar", "Saatnya Cek Gula Darah 🩸", "Jangan lupa cek dan catat gula darahmu hari ini, ya."},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			key := tt.key
			title, body := buildReminderNotification(domain.Reminder{SystemKey: &key})
			if title != tt.wantTitle || body != tt.wantBody {
				t.Fatalf("got (%q, %q), want (%q, %q)", title, body, tt.wantTitle, tt.wantBody)
			}
		})
	}
}

func TestDeduplicateDueRemindersPrefersSystemAndPreservesDistinctMedicine(t *testing.T) {
	day := []domain.ReminderActiveDay{{DayOfWeek: 2}}
	otherDay := []domain.ReminderActiveDay{{DayOfWeek: 6}}
	key := "weekly_blood_sugar"
	items := []domain.Reminder{
		{BaseModel: domain.BaseModel{ID: "personal"}, PatientID: "patient", ScheduledTime: "20:57:00", IsActive: true, ActionType: "blood_sugar_entry", ActiveDays: day},
		{BaseModel: domain.BaseModel{ID: "system"}, PatientID: "patient", ScheduledTime: "20:57:00", IsActive: true, ActionType: "blood_sugar_entry", SystemKey: &key, ActiveDays: day},
		{BaseModel: domain.BaseModel{ID: "medicine-a"}, PatientID: "patient", ScheduledTime: "20:57:00", IsActive: true, ActionType: "medication_entry", ActivityName: "Obat A", ActiveDays: day},
		{BaseModel: domain.BaseModel{ID: "medicine-b"}, PatientID: "patient", ScheduledTime: "20:57:00", IsActive: true, ActionType: "medication_entry", ActivityName: "Obat B", ActiveDays: day},
		{BaseModel: domain.BaseModel{ID: "inactive-day"}, PatientID: "patient", ScheduledTime: "20:57:00", IsActive: true, ActionType: "blood_sugar_entry", ActiveDays: otherDay},
	}
	selected := deduplicateDueReminders(items, 2)
	if len(selected) != 3 || selected[0].ID != "system" {
		t.Fatalf("selected=%v, want system and two distinct medicines", selected)
	}
}

func TestCatchUpMinutesIncludesScheduledTimes(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, now, want string }{
		{"meal", "2026-11-03 12:03", "2026-11-03 12:00"},
		{"exercise", "2026-11-03 16:04", "2026-11-03 16:00"},
		{"saturday blood sugar", "2026-11-07 07:05", "2026-11-07 07:00"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			now, err := time.ParseInLocation("2006-01-02 15:04", tc.now, loc)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, minute := range catchUpMinutes(now, 5) {
				if minute.Format("2006-01-02 15:04") == tc.want {
					found = true
				}
			}
			if !found {
				t.Fatalf("%s missing from catch-up window", tc.want)
			}
		})
	}
}
