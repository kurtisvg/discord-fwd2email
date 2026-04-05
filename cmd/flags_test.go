package cmd

import (
	"testing"
)

func TestParseFlags(t *testing.T) {
	t.Parallel()

	opts := parseFlags([]string{"-port", "9090"})
	if opts.port != "9090" {
		t.Fatalf("expected port 9090, got %s", opts.port)
	}
}

func TestParseFlags_Defaults(t *testing.T) {
	t.Parallel()

	opts := parseFlags([]string{})
	if opts.port != "8080" {
		t.Fatalf("expected default port 8080, got %s", opts.port)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    options
		wantErr bool
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
			wantErr: false,
		},
		{"missing token", options{discordAppID: "app", discordPublicKey: "key", gmailUser: "u", gmailAppPassword: "p"}, true},
		{"missing app id", options{discordToken: "tok", discordPublicKey: "key", gmailUser: "u", gmailAppPassword: "p"}, true},
		{"missing public key", options{discordToken: "tok", discordAppID: "app", gmailUser: "u", gmailAppPassword: "p"}, true},
		{"missing gmail user", options{discordToken: "tok", discordAppID: "app", discordPublicKey: "key", gmailAppPassword: "p"}, true},
		{"missing gmail password", options{discordToken: "tok", discordAppID: "app", discordPublicKey: "key", gmailUser: "u"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.opts.validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
