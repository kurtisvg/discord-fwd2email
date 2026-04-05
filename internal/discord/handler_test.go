package discord

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func newTestHandler(t *testing.T) (*Handler, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &Handler{publicKey: pub}, priv
}

func signBody(t *testing.T, priv ed25519.PrivateKey, timestamp string, body []byte) string {
	t.Helper()
	msg := make([]byte, 0, len(timestamp)+len(body))
	msg = append(msg, []byte(timestamp)...)
	msg = append(msg, body...)
	return hex.EncodeToString(ed25519.Sign(priv, msg))
}

func postInteraction(t *testing.T, h *Handler, sig, timestamp string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/interactions", bytes.NewReader(body))
	req.Header.Set("X-Signature-Ed25519", sig)
	req.Header.Set("X-Signature-Timestamp", timestamp)
	rec := httptest.NewRecorder()
	h.HandleInteraction(rec, req)
	return rec
}

func TestHandleInteraction_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	h, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/interactions", nil)
	rec := httptest.NewRecorder()
	h.HandleInteraction(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleInteraction_Ping(t *testing.T) {
	t.Parallel()
	h, priv := newTestHandler(t)

	body, _ := json.Marshal(discordgo.Interaction{Type: discordgo.InteractionPing})
	sig := signBody(t, priv, "1234567890", body)

	rec := postInteraction(t, h, sig, "1234567890", body)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp discordgo.InteractionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if resp.Type != discordgo.InteractionResponsePong {
		t.Fatalf("expected pong (type 1), got type %d", resp.Type)
	}
}

func TestHandleInteraction_InvalidSignature(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sig       string
		timestamp string
	}{
		{"malformed hex", "not-hex", "1234567890"},
		{"empty signature", "", "1234567890"},
		{"missing timestamp", "aabbccdd", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, _ := newTestHandler(t)
			body, _ := json.Marshal(discordgo.Interaction{Type: discordgo.InteractionPing})

			rec := postInteraction(t, h, tt.sig, tt.timestamp, body)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", rec.Code)
			}
		})
	}
}

func TestHandleInteraction_WrongKey(t *testing.T) {
	t.Parallel()
	h, _ := newTestHandler(t)

	// Sign with a different key.
	_, otherPriv, _ := ed25519.GenerateKey(nil)
	body, _ := json.Marshal(discordgo.Interaction{Type: discordgo.InteractionPing})
	sig := signBody(t, otherPriv, "1234567890", body)

	rec := postInteraction(t, h, sig, "1234567890", body)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestHandleInteraction_MalformedJSON(t *testing.T) {
	t.Parallel()
	h, priv := newTestHandler(t)

	body := []byte(`{not json}`)
	sig := signBody(t, priv, "1234567890", body)

	rec := postInteraction(t, h, sig, "1234567890", body)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestIsThread(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		chanType discordgo.ChannelType
		want     bool
	}{
		{
			name:     "public thread",
			chanType: discordgo.ChannelTypeGuildPublicThread,
			want:     true,
		},
		{
			name:     "private thread",
			chanType: discordgo.ChannelTypeGuildPrivateThread,
			want:     true,
		},
		{
			name:     "text channel",
			chanType: discordgo.ChannelTypeGuildText,
			want:     false,
		},
		{
			name:     "voice channel",
			chanType: discordgo.ChannelTypeGuildVoice,
			want:     false,
		},
		{
			name:     "DM",
			chanType: discordgo.ChannelTypeDM,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ch := &discordgo.Channel{Type: tt.chanType}
			if got := isThread(ch); got != tt.want {
				t.Fatalf("isThread(%d) = %v, want %v", tt.chanType, got, tt.want)
			}
		})
	}
}

func TestBuildSubject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		channelName string
		threadName  string
		serverName  string
		authorName  string
		want        string
	}{
		{
			name:        "channel only",
			channelName: "general",
			want:        "[Discord] Forwarded chat in #general",
		},
		{
			name:        "thread in channel",
			channelName: "support",
			threadName:  "billing issue",
			want:        "[Discord] Forwarded chat in #support › billing issue",
		},
		{
			name:       "DM",
			authorName: "Alice",
			want:       "[Discord] Forwarded DM with Alice",
		},
		{
			name: "fallback",
			want: "[Discord] Forwarded chat",
		},
		{
			name:        "channel with server (no thread)",
			channelName: "general",
			serverName:  "My Server",
			want:        "[Discord] Forwarded chat in #general",
		},
		{
			name:        "thread with server",
			channelName: "support",
			threadName:  "billing issue",
			serverName:  "My Server",
			want:        "[Discord] Forwarded chat in #support › billing issue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildSubject(tt.channelName, tt.threadName, tt.serverName, tt.authorName)
			if got != tt.want {
				t.Fatalf("buildSubject(%q, %q, %q, %q) = %q, want %q",
					tt.channelName, tt.threadName, tt.serverName, tt.authorName, got, tt.want)
			}
		})
	}
}
