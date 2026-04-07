package cmd

import (
	"strings"
	"testing"
)

func TestParseFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		wantPort string
	}{
		{"default port", []string{}, "8080"},
		{"custom port", []string{"-port", "9090"}, "9090"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			opts := parseFlags(tt.args)
			if opts.port != tt.wantPort {
				t.Fatalf("expected port %s, got %s", tt.wantPort, opts.port)
			}
		})
	}
}

func validSMTPOpts() options {
	return options{
		discordToken:     "tok",
		discordAppID:     "app",
		discordPublicKey: "key",
		emailProvider:    "smtp",
		toEmail:          "user@example.com",
		smtpUser:         "smtp@gmail.com",
		smtpPassword:     "pass",
	}
}

func validResendOpts() options {
	return options{
		discordToken:     "tok",
		discordAppID:     "app",
		discordPublicKey: "key",
		emailProvider:    "resend",
		toEmail:          "user@example.com",
		fromEmail:        "bot@example.com",
		resendAPIKey:     "re_xxx",
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		opts        options
		errContains string
	}{
		{
			name: "smtp valid",
			opts: validSMTPOpts(),
		},
		{
			name: "resend valid",
			opts: validResendOpts(),
		},
		{
			name:        "missing token",
			opts:        func() options { o := validSMTPOpts(); o.discordToken = ""; return o }(),
			errContains: "discord-token",
		},
		{
			name:        "missing app id",
			opts:        func() options { o := validSMTPOpts(); o.discordAppID = ""; return o }(),
			errContains: "discord-app-id",
		},
		{
			name:        "missing public key",
			opts:        func() options { o := validSMTPOpts(); o.discordPublicKey = ""; return o }(),
			errContains: "discord-public-key",
		},
		{
			name: "gateway skips public key",
			opts: func() options { o := validSMTPOpts(); o.discordPublicKey = ""; o.gateway = true; return o }(),
		},
		{
			name:        "missing to-email",
			opts:        func() options { o := validSMTPOpts(); o.toEmail = ""; return o }(),
			errContains: "to-email",
		},
		{
			name:        "smtp missing user",
			opts:        func() options { o := validSMTPOpts(); o.smtpUser = ""; return o }(),
			errContains: "smtp-user",
		},
		{
			name:        "smtp missing password",
			opts:        func() options { o := validSMTPOpts(); o.smtpPassword = ""; return o }(),
			errContains: "smtp-password",
		},
		{
			name:        "resend missing api key",
			opts:        func() options { o := validResendOpts(); o.resendAPIKey = ""; return o }(),
			errContains: "resend-api-key",
		},
		{
			name:        "resend missing from email",
			opts:        func() options { o := validResendOpts(); o.fromEmail = ""; return o }(),
			errContains: "from-email",
		},
		{
			name:        "unknown provider",
			opts:        func() options { o := validSMTPOpts(); o.emailProvider = "mailgun"; return o }(),
			errContains: "unknown email provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.opts.validate()
			if tt.errContains == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.errContains)
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}
