package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/container"
	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/infrastructure/notifications"
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

	repo := reminder.NewReminderRepository(c.DB, c.Logger)
	location, err := time.LoadLocation(c.Config.App.Timezone)
	if err != nil {
		log.Fatalf("failed to load application timezone: %v", err)
	}

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	run := func() {
		now := time.Now().In(location)
		if err := dispatchDueReminders(ctx, repo, sender, now); err != nil {
			c.Logger.Error("reminder dispatch failed", zap.Error(err))
		}
	}

	run()
	for range ticker.C {
		run()
	}
}

func buildReminderNotification(item domain.Reminder) (string, string) {
	icon := strings.ToLower(item.IconName)
	act := strings.TrimSpace(item.ActivityName)
	lowerAct := strings.ToLower(act)

	if item.Notes != "" {
		title := fmt.Sprintf("Waktunya %s ⏰", act)
		if act == "" {
			title = "Pengingat Rutinitas DIBA ⏰"
		}
		return title, item.Notes
	}

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
	case icon == "water" || strings.Contains(lowerAct, "air") || strings.Contains(lowerAct, "minum"):
		if act != "" {
			return fmt.Sprintf("Waktunya %s! 💧", act), fmt.Sprintf("Segelas air putih segar siap bantu metabolisme dan hidrasi tubuhmu tetap lancar 🌊")
		}
		return "Waktunya Minum Air Putih 💧", "Segelas air putih segar siap bantu metabolisme tubuh tetap lancar 🌊"
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
	sender *notifications.FCMSender,
	now time.Time,
) error {
	items, err := repo.FindDueReminders(ctx, now.Format("15:04"))
	if err != nil {
		return err
	}

	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	for _, item := range items {
		if !item.IsActive || !hasActiveDay(item.ActiveDays, weekday) {
			continue
		}
		tokens, err := repo.FindDeviceTokens(ctx, item.PatientID)
		if err != nil {
			return err
		}

		title, body := buildReminderNotification(item)

		for _, token := range tokens {
			_, err = sender.Send(ctx, token.Token, title, body, map[string]string{
				"type":          "reminder",
				"reminder_id":   item.ID,
				"icon_name":     item.IconName,
				"activity_name": item.ActivityName,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func hasActiveDay(days []domain.ReminderActiveDay, weekday int) bool {
	for _, day := range days {
		if day.DayOfWeek == weekday {
			return true
		}
	}
	return false
}
