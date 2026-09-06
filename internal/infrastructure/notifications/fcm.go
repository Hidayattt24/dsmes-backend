package notifications

import (
	"context"
	"encoding/json"
	"fmt"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FCMSender struct {
	client *messaging.Client
}

func NewFCMSender(ctx context.Context, credentialsJSON string) (*FCMSender, error) {
	if credentialsJSON == "" {
		return nil, fmt.Errorf("FCM_CREDENTIALS_JSON is not configured")
	}
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(credentialsJSON), &raw); err != nil {
		return nil, fmt.Errorf("invalid FCM_CREDENTIALS_JSON: %w", err)
	}
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON(raw))
	if err != nil {
		return nil, fmt.Errorf("initialize Firebase: %w", err)
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize Firebase Messaging: %w", err)
	}
	return &FCMSender{client: client}, nil
}

func (s *FCMSender) Send(ctx context.Context, token, title, body string, data map[string]string) (string, error) {
	return s.client.Send(ctx, &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ChannelID: "dsmes_reminders_channel",
			},
		},
	})
}
