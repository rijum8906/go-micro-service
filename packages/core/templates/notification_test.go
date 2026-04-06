package templates

import (
	"io/fs"
	"strings"
	"testing"
)

func TestNotificationTemplateFilesExist(t *testing.T) {
	for _, name := range NotificationNames() {
		for _, file := range NotificationFiles(name) {
			if _, err := fs.Stat(FS, file); err != nil {
				t.Fatalf("expected embedded notification template file %q to exist: %v", file, err)
			}
		}
	}
}

func TestNotificationTemplateValidate(t *testing.T) {
	for _, name := range NotificationNames() {
		if err := name.Validate(); err != nil {
			t.Fatalf("expected notification template %q to validate: %v", name, err)
		}
	}

	if err := NotificationTemplate("").Validate(); err == nil {
		t.Fatal("expected empty notification template to fail validation")
	}

	if err := NotificationTemplate("missing-template").Validate(); err == nil {
		t.Fatal("expected unknown notification template to fail validation")
	}
}

func TestRenderNotification(t *testing.T) {
	rendered, err := RenderNotification(NotificationTemplatePush, map[string]string{
		"title":      "Relay Notification",
		"body":       "You have a new message",
		"action_url": "https://example.com/inbox",
	})
	if err != nil {
		t.Fatalf("expected notification template render to succeed: %v", err)
	}

	if !strings.Contains(rendered.Title, "Relay Notification") {
		t.Fatalf("expected rendered notification title to include template title, got %q", rendered.Title)
	}

	if !strings.Contains(rendered.Body, "new message") {
		t.Fatalf("expected rendered notification body to include template body, got %q", rendered.Body)
	}

	if !strings.Contains(rendered.ActionURL, "https://example.com/inbox") {
		t.Fatalf("expected rendered action url to include template url, got %q", rendered.ActionURL)
	}
}

func TestRenderNotificationMissingData(t *testing.T) {
	_, err := RenderNotification(NotificationTemplatePush, map[string]string{
		"title": "Relay Notification",
	})
	if err == nil {
		t.Fatal("expected notification render to fail when required template data is missing")
	}
}
