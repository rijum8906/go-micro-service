package nats

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/templates"
)

func TestPublishEmailUsesMetaJobSubject(t *testing.T) {
	server := runTestServer(t)
	publisher := newTestClient(t, server.ClientURL())
	observer := newTestClient(t, server.ClientURL())

	sub, err := observer.Conn.SubscribeSync(dto.JobEmailVerification.String())
	if err != nil {
		t.Fatalf("expected email observer subscription to succeed: %v", err)
	}
	t.Cleanup(func() {
		_ = sub.Unsubscribe()
	})
	if err := observer.Conn.Flush(); err != nil {
		t.Fatalf("expected email observer subscription flush to succeed: %v", err)
	}

	message := testEmailMessage()
	if appErr := publisher.PublishEmail(message); appErr != nil {
		t.Fatalf("expected email publish to succeed: %v", appErr)
	}

	msg, err := sub.NextMsg(2 * time.Second)
	if err != nil {
		t.Fatalf("expected published email message to arrive: %v", err)
	}

	assertMetaJobSubject(t, msg.Data, dto.JobEmailVerification.String())
}

func TestPublishNotificationUsesTopLevelJobSubject(t *testing.T) {
	server := runTestServer(t)
	publisher := newTestClient(t, server.ClientURL())
	observer := newTestClient(t, server.ClientURL())

	sub, err := observer.Conn.SubscribeSync(dto.JobNotificationPush.String())
	if err != nil {
		t.Fatalf("expected notification observer subscription to succeed: %v", err)
	}
	t.Cleanup(func() {
		_ = sub.Unsubscribe()
	})
	if err := observer.Conn.Flush(); err != nil {
		t.Fatalf("expected notification observer subscription flush to succeed: %v", err)
	}

	message := testNotificationMessage()
	if appErr := publisher.PublishNotification(message); appErr != nil {
		t.Fatalf("expected notification publish to succeed: %v", appErr)
	}

	msg, err := sub.NextMsg(2 * time.Second)
	if err != nil {
		t.Fatalf("expected published notification message to arrive: %v", err)
	}

	assertTopLevelJobSubject(t, msg.Data, dto.JobNotificationPush.String())
}

func TestSubscribeEmailDecodesMessage(t *testing.T) {
	server := runTestServer(t)
	subscriber := newTestClient(t, server.ClientURL())
	publisher := newTestClient(t, server.ClientURL())

	received := make(chan dto.EmailMessage, 1)
	sub, appErr := subscriber.SubscribeEmail(dto.JobEmailVerification.String(), func(message dto.EmailMessage) {
		received <- message
	})
	if appErr != nil {
		t.Fatalf("expected email subscription to succeed: %v", appErr)
	}
	t.Cleanup(func() {
		_ = sub.Unsubscribe()
	})
	if err := subscriber.Conn.Flush(); err != nil {
		t.Fatalf("expected email subscription flush to succeed: %v", err)
	}

	expected := testEmailMessage()
	if appErr := publisher.PublishJSON(expected.Meta.JobSubject.String(), expected); appErr != nil {
		t.Fatalf("expected email publish json to succeed: %v", appErr)
	}

	select {
	case actual := <-received:
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("expected subscribed email message to match published message: %#v != %#v", actual, expected)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected subscribed email message to be delivered")
	}
}

func TestSubscribeNotificationDecodesMessage(t *testing.T) {
	server := runTestServer(t)
	subscriber := newTestClient(t, server.ClientURL())
	publisher := newTestClient(t, server.ClientURL())

	received := make(chan dto.NotificationMessage, 1)
	sub, appErr := subscriber.SubscribeNotification(dto.JobNotificationPush.String(), func(message dto.NotificationMessage) {
		received <- message
	})
	if appErr != nil {
		t.Fatalf("expected notification subscription to succeed: %v", appErr)
	}
	t.Cleanup(func() {
		_ = sub.Unsubscribe()
	})
	if err := subscriber.Conn.Flush(); err != nil {
		t.Fatalf("expected notification subscription flush to succeed: %v", err)
	}

	expected := testNotificationMessage()
	if appErr := publisher.PublishJSON(expected.JobSubject.String(), expected); appErr != nil {
		t.Fatalf("expected notification publish json to succeed: %v", appErr)
	}

	select {
	case actual := <-received:
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("expected subscribed notification message to match published message: %#v != %#v", actual, expected)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected subscribed notification message to be delivered")
	}
}

func runTestServer(t *testing.T) *natsserver.Server {
	t.Helper()

	server, err := natsserver.NewServer(&natsserver.Options{
		Host:   "127.0.0.1",
		Port:   -1,
		NoLog:  true,
		NoSigs: true,
	})
	if err != nil {
		t.Fatalf("expected test nats server to be created: %v", err)
	}

	go server.Start()
	if !server.ReadyForConnections(5 * time.Second) {
		server.Shutdown()
		t.Fatal("expected test nats server to be ready for connections")
	}

	t.Cleanup(func() {
		server.Shutdown()
	})

	return server
}

func newTestClient(t *testing.T, url string) *Client {
	t.Helper()

	client, appErr := Connect(context.Background(), Config{
		URL:        url,
		ClientName: t.Name(),
	})
	if appErr != nil {
		t.Fatalf("expected nats test client connection to succeed: %v", appErr)
	}

	t.Cleanup(func() {
		client.Close()
	})

	return client
}

func assertMetaJobSubject(t *testing.T, raw []byte, expected string) {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("expected published payload to be valid json: %v", err)
	}

	if _, ok := payload["job_subject"]; ok {
		t.Fatal("expected published payload to keep job_subject inside meta")
	}

	meta, ok := payload["meta"].(map[string]any)
	if !ok {
		t.Fatalf("expected published payload to include meta object: %#v", payload["meta"])
	}

	if actual, ok := meta["job_subject"].(string); !ok || actual != expected {
		t.Fatalf("expected meta.job_subject %q, got %#v", expected, meta["job_subject"])
	}
}

func assertTopLevelJobSubject(t *testing.T, raw []byte, expected string) {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("expected published payload to be valid json: %v", err)
	}

	if _, ok := payload["meta"]; ok {
		t.Fatal("expected published payload to keep job_subject at the top level")
	}

	if actual, ok := payload["job_subject"].(string); !ok || actual != expected {
		t.Fatalf("expected job_subject %q, got %#v", expected, payload["job_subject"])
	}
}

func testEmailMessage() dto.EmailMessage {
	return dto.EmailMessage{
		Meta: dto.EmailMetadata{
			JobSubject: dto.JobEmailVerification,
			Sender: dto.EmailSender{
				Email: "relay@example.com",
				Name:  "Relay",
			},
			Recipients: []dto.EmailRecipient{
				{Email: "user@example.com", Name: "User"},
			},
		},
		BodyContent: map[string]string{
			"app_name": "Relay",
		},
		Template: templates.EmailTemplateVerifyEmail,
	}
}

func testNotificationMessage() dto.NotificationMessage {
	return dto.NotificationMessage{
		JobSubject: dto.JobNotificationPush,
		Channel:    dto.NotificationChannelPush,
		Recipient: dto.NotificationRecipient{
			DeviceToken: "device-token-1",
		},
		BodyContent: map[string]string{
			"title":      "Relay",
			"body":       "Notification body",
			"action_url": "https://example.com/notifications/1",
		},
		Template: templates.NotificationTemplatePush,
	}
}
