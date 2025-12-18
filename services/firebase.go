package services

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var FirebaseApp *firebase.App
var FCMClient *messaging.Client

func InitFirebase() error {
	ctx := context.Background()

	serviceAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH")
	if serviceAccountPath == "" {
		serviceAccountPath = "firebase-service-account.json"
	}

	if _, err := os.Stat(serviceAccountPath); os.IsNotExist(err) {
		log.Printf("Warning: Firebase service account file not found at %s. Push notifications will be disabled.", serviceAccountPath)
		return nil
	}

	opt := option.WithCredentialsFile(serviceAccountPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return fmt.Errorf("error initializing Firebase app: %v", err)
	}

	FirebaseApp = app

	client, err := app.Messaging(ctx)
	if err != nil {
		return fmt.Errorf("error getting Messaging client: %v", err)
	}

	FCMClient = client
	log.Println("Firebase initialized successfully")
	return nil
}

func SendPushNotification(token, title, body string, data map[string]string) error {
	if FCMClient == nil {
		log.Println("FCM client not initialized, skipping push notification")
		return nil
	}

	if token == "" {
		return fmt.Errorf("FCM token is empty")
	}

	ctx := context.Background()

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
				Sound:       "default",
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
					Badge: func() *int { i := 1; return &i }(),
				},
			},
		},
	}

	response, err := FCMClient.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("error sending FCM message: %v", err)
	}

	log.Printf("Successfully sent FCM message: %s", response)
	return nil
}

func SendPushNotificationToMultiple(tokens []string, title, body string, data map[string]string) error {
	if FCMClient == nil {
		log.Println("FCM client not initialized, skipping push notification")
		return nil
	}

	if len(tokens) == 0 {
		return nil
	}

	ctx := context.Background()

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound:       "default",
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
					Badge: func() *int { i := 1; return &i }(),
				},
			},
		},
	}

	response, err := FCMClient.SendEachForMulticast(ctx, message)
	if err != nil {
		return fmt.Errorf("error sending FCM multicast message: %v", err)
	}

	log.Printf("Successfully sent FCM multicast: %d success, %d failures", response.SuccessCount, response.FailureCount)
	return nil
}
