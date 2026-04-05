package email

import (
	"bytes"
	"strings"
	"testing"
)

func TestEmailTemplate_ServerChannel(t *testing.T) {
	t.Parallel()

	data := ForwardData{
		ServerName:  "Acme Corp",
		ChannelName: "support",
		MessageLink: "https://discord.com/channels/123/456/789",
		TargetMessage: MessageData{
			AuthorName: "Alice",
			Content:    "Hello world",
		},
	}

	var buf bytes.Buffer
	if err := emailTemplate.Execute(&buf, data); err != nil {
		t.Fatalf("template error: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "Acme Corp") {
		t.Error("missing server name")
	}
	if !strings.Contains(html, "#support") {
		t.Error("missing channel name")
	}
	if !strings.Contains(html, "Alice") {
		t.Error("missing author name")
	}
	if !strings.Contains(html, "Hello world") {
		t.Error("missing message content")
	}
	if !strings.Contains(html, "https://discord.com/channels/123/456/789") {
		t.Error("missing message link")
	}
}

func TestEmailTemplate_DM(t *testing.T) {
	t.Parallel()

	data := ForwardData{
		MessageLink: "https://discord.com/channels/@me/456/789",
		TargetMessage: MessageData{
			AuthorName: "Bob",
			Content:    "DM content",
		},
	}

	var buf bytes.Buffer
	if err := emailTemplate.Execute(&buf, data); err != nil {
		t.Fatalf("template error: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "Forwarded DM with Bob") {
		t.Error("missing DM header")
	}
}

func TestEmailTemplate_WithContext(t *testing.T) {
	t.Parallel()

	data := ForwardData{
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
	}

	var buf bytes.Buffer
	if err := emailTemplate.Execute(&buf, data); err != nil {
		t.Fatalf("template error: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "First message") {
		t.Error("missing first context message")
	}
	if !strings.Contains(html, "Second message") {
		t.Error("missing second context message")
	}
	if !strings.Contains(html, "Target message") {
		t.Error("missing target message")
	}

	// Target should have blurple highlight, context should not.
	targetIdx := strings.Index(html, "Target message")
	blurpleIdx := strings.LastIndex(html[:targetIdx], "#5865F2")
	firstMsgIdx := strings.Index(html, "First message")

	if blurpleIdx < firstMsgIdx {
		t.Error("blurple highlight should only be on target message, not context")
	}
}

func TestEmailTemplate_NoContext(t *testing.T) {
	t.Parallel()

	data := ForwardData{
		ServerName:  "Test Server",
		ChannelName: "general",
		MessageLink: "https://discord.com/channels/1/2/3",
		TargetMessage: MessageData{
			AuthorName: "Alice",
			Content:    "Solo message",
		},
	}

	var buf bytes.Buffer
	if err := emailTemplate.Execute(&buf, data); err != nil {
		t.Fatalf("template error: %v", err)
	}

	if !strings.Contains(buf.String(), "Solo message") {
		t.Error("missing target message")
	}
}
