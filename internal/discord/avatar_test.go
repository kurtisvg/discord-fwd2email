package discord

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestAvatarURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		user     *discordgo.User
		wantHost string
	}{
		{
			name:     "custom avatar",
			user:     &discordgo.User{ID: "123", Avatar: "abc123"},
			wantHost: "cdn.discordapp.com/avatars/123/abc123",
		},
		{
			name:     "no avatar",
			user:     &discordgo.User{ID: "123", Avatar: ""},
			wantHost: "cdn.discordapp.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			url := avatarURL(tt.user)
			if !strings.Contains(url, tt.wantHost) {
				t.Errorf("expected URL containing %q, got %q", tt.wantHost, url)
			}
		})
	}
}

func TestMessageData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		msg        *discordgo.Message
		wantAuthor string
	}{
		{
			name: "uses global name",
			msg: &discordgo.Message{
				Author:  &discordgo.User{GlobalName: "Alice", Username: "alice123", Avatar: "abc"},
				Content: "hello",
			},
			wantAuthor: "Alice",
		},
		{
			name: "falls back to username",
			msg: &discordgo.Message{
				Author:  &discordgo.User{Username: "bob456", Avatar: "def"},
				Content: "world",
			},
			wantAuthor: "bob456",
		},
		{
			name: "with attachments",
			msg: &discordgo.Message{
				Author: &discordgo.User{Username: "charlie", Avatar: "ghi"},
				Attachments: []*discordgo.MessageAttachment{
					{Filename: "photo.png", URL: "https://cdn.example.com/photo.png", ContentType: "image/png"},
					{Filename: "doc.pdf", URL: "https://cdn.example.com/doc.pdf", ContentType: "application/pdf"},
				},
			},
			wantAuthor: "charlie",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			md := messageData(tt.msg)
			if md.AuthorName != tt.wantAuthor {
				t.Errorf("expected author %q, got %q", tt.wantAuthor, md.AuthorName)
			}
			if md.AvatarURL == "" {
				t.Error("expected non-empty avatar URL")
			}
			if md.Content != tt.msg.Content {
				t.Errorf("expected content %q, got %q", tt.msg.Content, md.Content)
			}
			if len(md.Attachments) != len(tt.msg.Attachments) {
				t.Fatalf("expected %d attachments, got %d", len(tt.msg.Attachments), len(md.Attachments))
			}
			for i, a := range md.Attachments {
				if a.Filename != tt.msg.Attachments[i].Filename {
					t.Errorf("attachment %d: expected filename %q, got %q", i, tt.msg.Attachments[i].Filename, a.Filename)
				}
			}
		})
	}
}
