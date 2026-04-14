package mailer

import (
	"bytes"
	"os"
	"slices"
	"strconv"
	"testing"
	"time"

	gomail "github.com/wneessen/go-mail"

	"github.com/rijum8906/relay/packages/core/apperror"
)

func TestConnectValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		config    Config
		wantCode  apperror.ErrorCode
		wantField string
	}{
		{
			name: "missing host",
			config: Config{
				Port: 587,
			},
			wantCode:  apperror.CodeValidation,
			wantField: "Host",
		},
		{
			name: "username without password",
			config: Config{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user",
			},
			wantCode:  apperror.CodeValidation,
			wantField: "Username",
		},
		{
			name: "conflicting tls modes",
			config: Config{
				Host:        "smtp.example.com",
				Port:        465,
				UseTLS:      true,
				UseStartTLS: true,
			},
			wantCode:  apperror.CodeValidation,
			wantField: "UseTLS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, err := Connect(tt.config)
			if client != nil {
				t.Fatalf("Connect() client = %v, want nil", client)
			}
			if err == nil {
				t.Fatal("Connect() error = nil, want validation error")
			}
			if err.Code != tt.wantCode {
				t.Fatalf("Connect() error code = %s, want %s", err.Code, tt.wantCode)
			}
			if !hasDetail(err, tt.wantField) {
				t.Fatalf("Connect() details = %#v, want field %q", err.Details, tt.wantField)
			}
		})
	}
}

func TestConnect(t *testing.T) {
	t.Parallel()
	host := os.Getenv("TEST_SMTP_HOST")
	portRaw := os.Getenv("TEST_SMTP_PORT")
	portInt, err := strconv.Atoi(portRaw)
	if err != nil {
		panic(err)
	}
	username := os.Getenv("TEST_SMTP_USERNAME")
	password := os.Getenv("TEST_SMTP_PASSWORD")

	cfg := Config{
		Host:        host,
		Port:        portInt,
		Username:    username,
		Password:    password,
		FromEmail:   username,
		FromName:    "Riju Mondal",
		UseStartTLS: true,
		UseTLS:      false,
		Retries:     3,
		Timeout:     5 * time.Second,
	}

	_, appErr := Connect(cfg)
	if appErr != nil {
		panic(appErr)
	}
}

func TestSendValidation(t *testing.T) {
	t.Parallel()

	validMessage := testMessage()

	t.Run("nil client", func(t *testing.T) {
		t.Parallel()

		err := Send(nil, validMessage)
		if err == nil {
			t.Fatal("Send() error = nil, want error")
		}
		if err.Code != apperror.CodeInternal {
			t.Fatalf("Send() error code = %s, want %s", err.Code, apperror.CodeInternal)
		}
	})

	t.Run("invalid message", func(t *testing.T) {
		t.Parallel()

		err := Send(&gomail.Client{}, Message{})
		if err == nil {
			t.Fatal("Send() error = nil, want error")
		}
		if err.Code != apperror.CodeValidation {
			t.Fatalf("Send() error code = %s, want %s", err.Code, apperror.CodeValidation)
		}
	})
}

func TestSendWithConfigValidation(t *testing.T) {
	t.Parallel()

	err := SendWithConfig(Config{}, testMessage())
	if err == nil {
		t.Fatal("SendWithConfig() error = nil, want validation error")
	}
	if err.Code != apperror.CodeValidation {
		t.Fatalf("SendWithConfig() error code = %s, want %s", err.Code, apperror.CodeValidation)
	}
}

func TestBuildMessagePopulatesHeadersAndBody(t *testing.T) {
	t.Parallel()

	msg, err := buildMessage(testMessage())
	if err != nil {
		t.Fatalf("buildMessage() error = %v", err)
	}

	if got := msg.GetAddrHeader(gomail.HeaderFrom); len(got) != 1 || got[0].Address != "sender@example.com" {
		t.Fatalf("From header = %#v, want sender@example.com", got)
	}
	if got := msg.GetAddrHeader(gomail.HeaderTo); len(got) != 1 || got[0].Address != "to@example.com" {
		t.Fatalf("To header = %#v, want to@example.com", got)
	}
	if got := msg.GetAddrHeader(gomail.HeaderCc); len(got) != 1 || got[0].Address != "cc@example.com" {
		t.Fatalf("CC header = %#v, want cc@example.com", got)
	}
	if got := msg.GetAddrHeader(gomail.HeaderBcc); len(got) != 1 || got[0].Address != "bcc@example.com" {
		t.Fatalf("BCC header = %#v, want bcc@example.com", got)
	}

	if got := msg.GetGenHeader(gomail.HeaderSubject); !slices.Equal(got, []string{"Subject line"}) {
		t.Fatalf("Subject header = %#v, want %#v", got, []string{"Subject line"})
	}
	if got := msg.GetGenHeader(gomail.Header("X-Trace-ID")); !slices.Equal(got, []string{"abc-123"}) {
		t.Fatalf("X-Trace-ID header = %#v, want %#v", got, []string{"abc-123"})
	}
	if got := msg.GetGenHeader(gomail.HeaderImportance); !slices.Equal(got, []string{"high"}) {
		t.Fatalf("Importance header = %#v, want %#v", got, []string{"high"})
	}

	var rendered bytes.Buffer
	if _, writeErr := msg.WriteTo(&rendered); writeErr != nil {
		t.Fatalf("WriteTo() error = %v", writeErr)
	}

	output := rendered.String()
	for _, want := range []string{
		"plain body",
		"<p>html body</p>",
		`filename="report.txt"`,
		"text/plain",
	} {
		if !bytes.Contains(rendered.Bytes(), []byte(want)) {
			t.Fatalf("rendered message missing %q\n%s", want, output)
		}
	}
}

func TestBuildMessageRejectsWhitespaceOnlyBody(t *testing.T) {
	t.Parallel()

	_, err := buildMessage(Message{
		Envelope: Envelope{
			From: "sender@example.com",
			To:   []string{"to@example.com"},
		},
		Content: Content{
			Subject: "Subject line",
			Text:    "   \n\t  ",
			HTML:    "   ",
		},
	})
	if err == nil {
		t.Fatal("buildMessage() error = nil, want validation error")
	}
	if err.Code != apperror.CodeValidation {
		t.Fatalf("buildMessage() error code = %s, want %s", err.Code, apperror.CodeValidation)
	}
}

func TestMailAddresses(t *testing.T) {
	t.Parallel()

	if got, err := mailAddresses(nil); err != nil || got != nil {
		t.Fatalf("mailAddresses(nil) = %#v, want nil", got)
	}

	addresses := []string{
		"One <one@example.com>",
		"Two <two@example.com>",
	}
	got, err := mailAddresses(addresses)
	if err != nil {
		t.Fatalf("mailAddresses() error = %v", err)
	}
	if len(got) != len(addresses) {
		t.Fatalf("mailAddresses() len = %d, want %d", len(got), len(addresses))
	}
	if got[0].Name != "One" || got[0].Address != "one@example.com" {
		t.Fatalf("mailAddresses()[0] = %#v, want parsed first address", got[0])
	}
	if got[1].Name != "Two" || got[1].Address != "two@example.com" {
		t.Fatalf("mailAddresses()[1] = %#v, want parsed second address", got[1])
	}
}

func TestToImportance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		priority EmailPriority
		want     gomail.Importance
	}{
		{name: "low", priority: EmailPriorityLow, want: gomail.ImportanceLow},
		{name: "normal", priority: EmailPriorityNormal, want: gomail.ImportanceNormal},
		{name: "high", priority: EmailPriorityHigh, want: gomail.ImportanceHigh},
		{name: "default", priority: EmailPriority(99), want: gomail.ImportanceNormal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			t.Parallel()
			if got := toImportance(tt.priority); got != tt.want {
				t.Fatalf("toImportance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConnectContext(t *testing.T) {
	t.Parallel()

	t.Run("no timeout", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := connectContext(0)
		defer cancel()

		if _, ok := ctx.Deadline(); ok {
			t.Fatal("connectContext(0) set a deadline, want none")
		}
	})

	t.Run("with timeout", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := connectContext(50 * time.Millisecond)
		defer cancel()

		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("connectContext(timeout) deadline missing")
		}

		remaining := time.Until(deadline)
		if remaining <= 0 || remaining > 100*time.Millisecond {
			t.Fatalf("connectContext(timeout) deadline = %v from now, want within (0, 100ms]", remaining)
		}
	})
}

func testMessage() Message {
	return Message{
		Envelope: Envelope{
			From: "Sender <sender@example.com>",
			To:   []string{"To <to@example.com>"},
			CC:   []string{"CC <cc@example.com>"},
			BCC:  []string{"BCC <bcc@example.com>"},
		},
		Content: Content{
			Subject:  "Subject line",
			Text:     "plain body",
			HTML:     "<p>html body</p>",
			Priority: EmailPriorityHigh,
			Headers: map[string]string{
				"X-Trace-ID": "abc-123",
			},
			Attachments: []Attachment{
				{
					Filename:    "report.txt",
					ContentType: "text/plain",
					Data:        []byte("report body"),
				},
			},
		},
	}
}

func hasDetail(err *apperror.AppError, field string) bool {
	for _, detail := range err.Details {
		if detail.Field == field {
			return true
		}
	}
	return false
}
