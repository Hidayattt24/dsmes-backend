package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/container"
	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/infrastructure/notifications"
	"github.com/dsmes/dsmes-backend/internal/modules/education"
	"github.com/dsmes/dsmes-backend/internal/modules/reminder"
)

func main() {
	c, err := container.Build()
	if err != nil {
		log.Fatalf("failed to build worker container: %v", err)
	}
	defer c.Close()

	ctx := context.Background()
	sender, err := notifications.NewFCMSender(ctx, c.Config.FCM.CredentialsJSON)
	if err != nil {
		log.Fatalf("failed to initialize FCM sender: %v", err)
	}
	c.Logger.Info("reminder FCM sender ready")

	repo := reminder.NewReminderRepository(c.DB, c.Logger)
	educationFollowups := education.NewEducationFollowupRepository(c.DB)
	location, err := time.LoadLocation(c.Config.App.Timezone)
	if err != nil {
		log.Fatalf("failed to load application timezone: %v", err)
	}

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	run := func() {
		now := time.Now().In(location)
		for _, dueAt := range catchUpMinutes(now, 5) {
			if err := dispatchDueReminders(ctx, repo, sender, dueAt, c.Logger); err != nil {
				c.Logger.Error("reminder dispatch query failed", zap.Time("scheduled_at", dueAt), zap.Error(err))
			}
		}
		if err := education.DispatchDueEducationFollowups(ctx, educationFollowups, sender, now, c.Logger); err != nil {
			c.Logger.Error("education followup dispatch failed", zap.Error(err))
		}
	}

	run()
	for range ticker.C {
		run()
	}
}

func catchUpMinutes(now time.Time, previousMinutes int) []time.Time {
	current := now.Truncate(time.Minute)
	minutes := make([]time.Time, 0, previousMinutes+1)
	for offset := previousMinutes; offset >= 0; offset-- {
		minutes = append(minutes, current.Add(-time.Duration(offset)*time.Minute))
	}
	return minutes
}

func buildReminderNotification(item domain.Reminder) (string, string) {
	if item.SystemKey != nil {
		switch *item.SystemKey {
		case "daily_meal":
			return "Waktunya Catat Makanan 🥗", "Jangan lupa mencatat makanan yang kamu konsumsi hari ini, ya."
		case "daily_exercise":
			return "Waktunya Bergerak 🏃", "Yuk luangkan waktu untuk berolahraga dan tetap aktif sore ini."
		case "weekly_blood_sugar":
			return "Saatnya Cek Gula Darah 🩸", "Jangan lupa cek dan catat gula darahmu hari ini, ya."
		}
	}
	icon := strings.ToLower(item.IconName)
	act := strings.TrimSpace(item.ActivityName)
	lowerAct := strings.ToLower(act)

	switch {
	case icon == "blood" || strings.Contains(lowerAct, "gula") || strings.Contains(lowerAct, "glukosa"):
		if act != "" {
			return fmt.Sprintf("Waktunya %s! 🩸", act), fmt.Sprintf("Saatnya mencatat angka %s demi memantau kesehatan tubuhmu tetap prima ✨", act)
		}
		return "Waktunya Cek Kadar Gula Darah! 🩸", "Luangkan 1 menit untuk mencatat angka gula darahmu demi tubuh tetap prima ✨"
	case icon == "medicine" || strings.Contains(lowerAct, "obat") || strings.Contains(lowerAct, "insulin"):
		if act != "" {
			return fmt.Sprintf("Waktunya %s! 💊", act), fmt.Sprintf("Saatnya konsumsi %s tepat waktu sesuai anjuran dokter agar tubuh tetap stabil 💪", act)
		}
		return "Waktunya Konsumsi Obat / Insulin 💊", "Jaga kestabilan tubuh dengan minum obat tepat waktu sesuai anjuran dokter 💪"
	case icon == "walk" || icon == "run" || icon == "exercise" || icon == "bike" || icon == "yoga" || strings.Contains(lowerAct, "jalan") || strings.Contains(lowerAct, "olahraga") || strings.Contains(lowerAct, "senam"):
		if act != "" {
			return fmt.Sprintf("Waktunya %s! 🏃‍♂️", act), fmt.Sprintf("Yuk lakukan %s sejenak untuk membantu menjaga kadar gula darah tetap stabil 🌟", act)
		}
		return "Waktunya Bergerak & Segarkan Tubuh! 🏃‍♂️", "Jalan santai sejenak sangat membantu menjaga kadar gula darah tetap stabil 🌟"
	case icon == "breakfast" || icon == "lunch" || icon == "dinner" || icon == "restaurant" || strings.Contains(lowerAct, "makan") || strings.Contains(lowerAct, "sarapan"):
		if act != "" {
			return fmt.Sprintf("Waktunya %s! 🥗", act), fmt.Sprintf("Saatnya %s. Penuhi nutrisi seimbang sesuai takaran piring sehatmu hari ini 🍏", act)
		}
		return "Waktunya Jadwal Makan Sehat 🥗", "Penuhi nutrisi seimbang sesuai takaran piring sehatmu hari ini 🍏"
	case icon == "water" || icon == "coffee" || strings.Contains(lowerAct, "air") || strings.Contains(lowerAct, "minum"):
		if act != "" {
			return fmt.Sprintf("Waktunya %s! 💧", act), fmt.Sprintf("Segelas air putih segar siap bantu metabolisme dan hidrasi tubuhmu tetap lancar 🌊")
		}
		return "Waktunya Minum Air Putih 💧", "Segelas air putih segar siap bantu metabolisme tubuh tetap lancar 🌊"
	case icon == "alarm" || icon == "clock" || icon == "work" || icon == "home":
		if act != "" {
			return fmt.Sprintf("Waktunya %s! ⏰", act), "Yuk lakukan rutinitasmu sesuai jadwal agar tetap konsisten hari ini ✨"
		}
		return "Waktunya Menjalankan Rutinitas! ⏰", "Yuk lakukan rutinitasmu sesuai jadwal agar tetap konsisten hari ini ✨"
	case icon == "sleep" || strings.Contains(lowerAct, "tidur") || strings.Contains(lowerAct, "istirahat"):
		if act != "" {
			return fmt.Sprintf("Waktunya %s! 🌙", act), fmt.Sprintf("Tidur berkualitas dan %s sangat penting untuk pemulihan metabolisme tubuhmu 😴", act)
		}
		return "Waktunya Istirahat yang Cukup 🌙", "Tidur berkualitas sangat penting untuk pemulihan dan kestabilan metabolisme tubuhmu 😴"
	case icon == "hospital" || strings.Contains(lowerAct, "kontrol") || strings.Contains(lowerAct, "dokter") || strings.Contains(lowerAct, "puskesmas"):
		if act != "" {
			return fmt.Sprintf("Jadwal %s 🏥", act), fmt.Sprintf("Jangan lupa jadwal %s untuk pemantauan kesehatan yang optimal 🩺", act)
		}
		return "Jadwal Pemeriksaan Kesehatan 🏥", "Jangan lupa jadwal konsultasi/kontrol kesehatanmu untuk pemantauan yang optimal 🩺"
	case icon == "heart" || strings.Contains(lowerAct, "jantung"):
		if act != "" {
			return fmt.Sprintf("Waktunya %s! ❤️", act), fmt.Sprintf("Yuk lakukan %s untuk menjaga detak jantung dan tekanan darah tetap stabil 🧘", act)
		}
		return "Jaga Kesehatan Jantung & Tubuhmu ❤️", "Yuk lakukan rileksasi sejenak untuk menjaga detak jantung dan tekanan darah tetap stabil 🧘"
	default:
		if act != "" {
			return fmt.Sprintf("Waktunya %s ⏰", act), fmt.Sprintf("Halo! Saatnya melakukan %s. Yuk jaga konsistensi rutinitas sehatmu hari ini! ✨", act)
		}
		return "Pengingat Rutinitas DIBA ⏰", "Yuk terus jaga konsistensi rutinitas kesehatanmu hari ini! ✨"
	}
}

func dispatchDueReminders(
	ctx context.Context,
	repo reminder.ReminderRepository,
	sender interface {
		Send(context.Context, string, string, string, map[string]string) (string, error)
	},
	now time.Time,
	logger *zap.Logger,
) error {
	items, err := repo.FindDueReminders(ctx, now.Format("15:04"))
	if err != nil {
		return err
	}

	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	items = deduplicateDueReminders(items, weekday)
	var due, stored, targets, sent, failed int
	for _, item := range items {
		if !item.IsActive || !hasActiveDay(item.ActiveDays, weekday) {
			continue
		}
		due++
		title, body := buildReminderNotification(item)
		if item.NotificationName != "" {
			body = fmt.Sprintf("Hai, %s! %s", item.NotificationName, body)
		}
		scheduledAt := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location()).UTC()
		occurrenceID := fmt.Sprintf("%s:%s", item.ID, scheduledAt.Format(time.RFC3339))
		created, err := repo.CreateNotificationOccurrence(ctx, &domain.NotificationLog{
			ReminderID: &item.ID, PatientID: item.PatientID, MessageText: body,
			NotifiedAt: now.UTC(), NotifType: "reminder", ActionType: item.ActionType,
			OccurrenceID: &occurrenceID, ScheduledAt: &scheduledAt,
		})
		if err != nil {
			logger.Warn("failed to store reminder occurrence", zap.String("reminder_id", item.ID), zap.Error(err))
			continue
		}
		if !created {
			continue
		}
		stored++

		tokens, err := repo.FindDeviceTokens(ctx, item.PatientID)
		if err != nil {
			logger.Warn("failed to fetch device tokens for patient", zap.String("patient_id", item.PatientID), zap.Error(err))
			continue
		}

		seenTokens := make(map[string]bool)
		var itemTargets, itemSent, itemFailed int
		sentAtStr := time.Now().UTC().Format(time.RFC3339)
		scheduledAtStr := scheduledAt.Format(time.RFC3339)
		for _, token := range tokens {
			if seenTokens[token.Token] {
				continue
			}
			seenTokens[token.Token] = true
			targets++
			itemTargets++

			_, err = sender.Send(ctx, token.Token, title, body, map[string]string{
				"type":          "reminder",
				"reminder_id":   item.ID,
				"icon_name":     item.IconName,
				"activity_name": item.ActivityName,
				"category":      string(item.Category),
				"action_type":   item.ActionType,
				"occurrence_id": occurrenceID,
				"scheduled_at":  scheduledAtStr,
				"sent_at":       sentAtStr,
			})
			if err != nil {
				failed++
				itemFailed++
				if notifications.IsUnregistered(err) {
					logger.Info("deleting stale device token",
						zap.String("patient_id", item.PatientID),
					)
					_ = repo.DeleteDeviceToken(ctx, item.PatientID, token.Token)
				} else {
					logger.Warn("failed to send reminder notification",
						zap.String("patient_id", item.PatientID),
						zap.String("reminder_id", item.ID),
						zap.Error(err),
					)
				}
			} else {
				sent++
				itemSent++
			}
		}
		logger.Info("reminder occurrence dispatched",
			zap.String("occurrence_id", occurrenceID),
			zap.Int("target_tokens", itemTargets),
			zap.Int("sent", itemSent),
			zap.Int("failed", itemFailed),
		)
	}
	fields := []zap.Field{zap.Time("scheduled_at", now), zap.Int("due", due), zap.Int("inbox_created", stored), zap.Int("device_targets", targets), zap.Int("push_sent", sent), zap.Int("push_failed", failed)}
	if due > 0 {
		logger.Info("reminder dispatch completed", fields...)
	} else {
		logger.Debug("reminder dispatch scan completed", fields...)
	}
	return nil
}

// Prefer the system reminder when a patient has a personal reminder for the
// same action at the same scheduled minute. Medication and generic reminders
// retain their activity names so distinct medicines/routines are not merged.
func deduplicateDueReminders(items []domain.Reminder, weekday int) []domain.Reminder {
	sort.SliceStable(items, func(i, j int) bool {
		if (items[i].SystemKey != nil) != (items[j].SystemKey != nil) {
			return items[i].SystemKey != nil
		}
		return items[i].ID < items[j].ID
	})
	seen := make(map[string]bool, len(items))
	systemActions := make(map[string]bool, len(items))
	for _, item := range items {
		if item.SystemKey != nil && item.IsActive && hasActiveDay(item.ActiveDays, weekday) {
			minute := item.ScheduledTime
			if len(minute) >= 5 {
				minute = minute[:5]
			}
			systemActions[item.PatientID+"|"+minute+"|"+item.ActionType] = true
		}
	}
	result := make([]domain.Reminder, 0, len(items))
	for _, item := range items {
		if !item.IsActive || !hasActiveDay(item.ActiveDays, weekday) {
			continue
		}
		action := item.ActionType
		if action == "" {
			action = item.ID
		}
		minute := item.ScheduledTime
		if len(minute) >= 5 {
			minute = minute[:5]
		}
		baseKey := item.PatientID + "|" + minute + "|" + action
		if item.SystemKey == nil && systemActions[baseKey] &&
			action != "medication_entry" && action != "reminder_management" {
			continue
		}
		key := baseKey
		if item.SystemKey == nil {
			key += "|" + strings.ToLower(strings.TrimSpace(item.ActivityName))
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, item)
	}
	return result
}

func hasActiveDay(days []domain.ReminderActiveDay, weekday int) bool {
	for _, day := range days {
		if day.DayOfWeek == weekday {
			return true
		}
	}
	return false
}
