package email

import (
	"bytes"
	"strings"
	"testing"
)

func TestEmailTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		data       ForwardData
		want       []string
		wantAbsent []string
	}{
		{
			name: "server and channel",
			data: ForwardData{
				ServerName:  "Acme Corp",
				ChannelName: "support",
				MessageLink: "https://discord.com/channels/123/456/789",
				TargetMessage: MessageData{
					AuthorName: "Alice",
					Content:    "Hello world",
				},
			},
			want: []string{
				"Acme Corp",
				"#support",
				"Alice",
				"Hello world",
				"https://discord.com/channels/123/456/789",
				"Open in Discord",
			},
			wantAbsent: []string{"Forwarded DM"},
		},
		{
			name: "DM with no server",
			data: ForwardData{
				IsDM:        true,
				MessageLink: "https://discord.com/channels/@me/456/789",
				TargetMessage: MessageData{
					AuthorName: "Bob",
					Content:    "DM content",
				},
			},
			want:       []string{"Forwarded DM with Bob", "DM content"},
			wantAbsent: []string{"Acme Corp", "#support"},
		},
		{
			name: "with context messages",
			data: ForwardData{
				ServerName:  "Test Server",
				ChannelName: "general",
				MessageLink: "https://discord.com/channels/1/2/3",
				ContextMessages: []MessageData{
					{AuthorName: "Alice", Content: "First message"},
					{AuthorName: "Bob", Content: "Second message"},
				},
				TargetMessage: MessageData{
					AuthorName: "Charlie",
					Content:    "Target message",
				},
			},
			want: []string{"First message", "Second message", "Target message"},
		},
		{
			name: "no context messages",
			data: ForwardData{
				ServerName:  "Test Server",
				ChannelName: "general",
				MessageLink: "https://discord.com/channels/1/2/3",
				TargetMessage: MessageData{
					AuthorName: "Alice",
					Content:    "Solo message",
				},
			},
			want: []string{"Solo message", "Alice"},
		},
		{
			name: "image attachment renders inline",
			data: ForwardData{
				ServerName:  "Test",
				ChannelName: "general",
				MessageLink: "https://discord.com/channels/1/2/3",
				TargetMessage: MessageData{
					AuthorName: "Alice",
					Content:    "Check this out",
					Attachments: []Attachment{
						{Filename: "photo.png", URL: "https://cdn.example.com/photo.png", IsImage: true},
					},
				},
			},
			want: []string{"Check this out", `<img src="https://cdn.example.com/photo.png"`},
		},
		{
			name: "file attachment renders as link",
			data: ForwardData{
				ServerName:  "Test",
				ChannelName: "general",
				MessageLink: "https://discord.com/channels/1/2/3",
				TargetMessage: MessageData{
					AuthorName: "Bob",
					Content:    "Here's the doc",
					Attachments: []Attachment{
						{Filename: "report.pdf", URL: "https://cdn.example.com/report.pdf", IsImage: false},
					},
				},
			},
			want: []string{"report.pdf", `href="https://cdn.example.com/report.pdf"`},
		},
		{
			name: "thread in server",
			data: ForwardData{
				ServerName:  "Acme Corp",
				ChannelName: "support",
				ThreadName:  "billing issue",
				MessageLink: "https://discord.com/channels/123/456/789",
				TargetMessage: MessageData{
					AuthorName: "Alice",
					Content:    "Thread content",
				},
			},
			want: []string{
				"Acme Corp",
				"#support",
				"› billing issue",
				"Thread content",
			},
			wantAbsent: []string{"Forwarded DM"},
		},
		{
			name: "thread without server",
			data: ForwardData{
				ChannelName: "general",
				ThreadName:  "my thread",
				MessageLink: "https://discord.com/channels/123/456/789",
				TargetMessage: MessageData{
					AuthorName: "Bob",
					Content:    "Hello",
				},
			},
			want: []string{
				"#general",
				"› my thread",
			},
			wantAbsent: []string{"Forwarded DM"},
		},
		{
			name: "attachment only message (no text)",
			data: ForwardData{
				ServerName:  "Test",
				ChannelName: "general",
				MessageLink: "https://discord.com/channels/1/2/3",
				TargetMessage: MessageData{
					AuthorName: "Charlie",
					Attachments: []Attachment{
						{Filename: "image.jpg", URL: "https://cdn.example.com/image.jpg", IsImage: true},
					},
				},
			},
			want:       []string{"Charlie", `<img src="https://cdn.example.com/image.jpg"`},
			wantAbsent: []string{"<p style=\"margin:0;font-size:14px"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			if err := emailTemplate.Execute(&buf, tt.data); err != nil {
				t.Fatalf("template error: %v", err)
			}
			html := buf.String()

			for _, s := range tt.want {
				if !strings.Contains(html, s) {
					t.Errorf("expected %q in output", s)
				}
			}
			for _, s := range tt.wantAbsent {
				if strings.Contains(html, s) {
					t.Errorf("did not expect %q in output", s)
				}
			}
		})
	}
}

func TestEmailTemplate_TargetHighlight(t *testing.T) {
	t.Parallel()

	data := ForwardData{
		ServerName:  "Test",
		ChannelName: "general",
		MessageLink: "https://discord.com/channels/1/2/3",
		ContextMessages: []MessageData{
			{AuthorName: "Alice", Content: "Context msg"},
		},
		TargetMessage: MessageData{
			AuthorName: "Bob",
			Content:    "Target msg",
		},
	}

	var buf bytes.Buffer
	if err := emailTemplate.Execute(&buf, data); err != nil {
		t.Fatalf("template error: %v", err)
	}
	html := buf.String()

	// The context message table should not have the tinted background.
	contextIdx := strings.Index(html, "Context msg")
	contextTable := html[strings.LastIndex(html[:contextIdx], "<table"):contextIdx]

	targetIdx := strings.Index(html, "Target msg")
	targetTable := html[strings.LastIndex(html[:targetIdx], "<table"):targetIdx]

	if strings.Contains(contextTable, "background-color:#f0f4ff") {
		t.Error("context message should not have tinted background")
	}
	if !strings.Contains(targetTable, "background-color:#f0f4ff") {
		t.Error("target message should have tinted background")
	}
}
