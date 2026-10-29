package fcm

import (
	"context"
	"errors"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
)

type Service struct {
	client *messaging.Client
}

func NewService(ctx context.Context) (*Service, error) {
	projectID := strings.TrimSpace(
		os.Getenv("FIREBASE_PROJECT_ID"),
	)

	if projectID == "" {
		projectID = "helpdesk-nutech"
	}

	app, err := firebase.NewApp(
		ctx,
		&firebase.Config{
			ProjectID: projectID,
		},
	)
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return &Service{
		client: client,
	}, nil
}

func (s *Service) SendToToken(ctx context.Context, token string, title string, body string, data map[string]string) error {
	if s == nil || s.client == nil {
		return errors.New("Firebase Messaging belum diinisialisasi")
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("FCM registration token kosong")
	}

	if strings.TrimSpace(title) == "" {
		title = "Notifikasi Helpdesk"
	}

	if strings.TrimSpace(body) == "" {
		body = "Ada pembaruan tiket."
	}

	message := &messaging.Message{
		Token: token,

		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},

		Data: data,

		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ChannelID: "helpdesk_updates",
			},
		},
	}

	_, err := s.client.Send(ctx, message)
	return err
}
