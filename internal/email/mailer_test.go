package email

import (
	"strings"
	"testing"
)

func TestBuildMIME(t *testing.T) {
	t.Parallel()

	msg := string(buildMIME("from@test.com", "to@test.com", "Test Subject", "<h1>Hello</h1>"))

	wantHeaders := []string{
		"From: from@test.com",
		"To: to@test.com",
		"Subject: Test Subject",
		"MIME-Version: 1.0",
		`Content-Type: text/html; charset="UTF-8"`,
	}

	for _, header := range wantHeaders {
		if !strings.Contains(msg, header) {
			t.Errorf("missing header: %s", header)
		}
	}

	if !strings.Contains(msg, "<h1>Hello</h1>") {
		t.Error("missing HTML body")
	}

	if !strings.Contains(msg, "\r\n\r\n") {
		t.Error("missing blank line between headers and body")
	}
}
