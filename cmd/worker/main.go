package main

import (
	"context"
	"log"
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
		for _, token := range tokens {
			body := item.Notes
			if body == "" {
				body = "Waktunya melakukan " + item.ActivityName + "."
			}
			_, err = sender.Send(ctx, token.Token, "Pengingat DSMES", body, map[string]string{
				"type":        "reminder",
				"reminder_id": item.ID,
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
