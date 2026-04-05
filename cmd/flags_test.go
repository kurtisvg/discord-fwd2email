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

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		opts        options
		errContains string
	}{
		{
			name: "all set",
			opts: options{
				discordToken:     "tok",
				discordAppID:     "app",
				discordPublicKey: "key",
				gmailUser:        "user@gmail.com",
				gmailAppPassword: "pass",
			},
		},
		{
			name:        "missing token",
			opts:        options{discordAppID: "app", discordPublicKey: "key", gmailUser: "u", gmailAppPassword: "p"},
			errContains: "discord-token",
		},
		{
			name:        "missing app id",
			opts:        options{discordToken: "tok", discordPublicKey: "key", gmailUser: "u", gmailAppPassword: "p"},
			errContains: "discord-app-id",
		},
		{
			name:        "missing public key",
			opts:        options{discordToken: "tok", discordAppID: "app", gmailUser: "u", gmailAppPassword: "p"},
			errContains: "discord-public-key",
		},
		{
			name:        "missing gmail user",
			opts:        options{discordToken: "tok", discordAppID: "app", discordPublicKey: "key", gmailAppPassword: "p"},
			errContains: "gmail-user",
		},
		{
			name:        "missing gmail password",
			opts:        options{discordToken: "tok", discordAppID: "app", discordPublicKey: "key", gmailUser: "u"},
			errContains: "gmail-app-password",
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
