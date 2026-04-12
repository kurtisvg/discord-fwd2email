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

	// Snowflake 1360000000000000000 → Apr 10, 2025 9:14 PM UTC
	const msgID = "1360000000000000000"
	const ts = "Apr 10, 9:14 PM"

	tests := []struct {
		name        string
		channelName string
		threadName  string
		isDM        bool
		authorName  string
		want        string
	}{
		{
			name:        "channel only",
			channelName: "general",
			authorName:  "Bob",
			want:        "[Discord] Bob in #general — " + ts,
		},
		{
			name:        "thread in channel",
			channelName: "support",
			threadName:  "billing issue",
			authorName:  "Bob",
			want:        "[Discord] Bob in #support › billing issue — " + ts,
		},
		{
			name:       "DM",
			isDM:       true,
			authorName: "Alice",
			want:       "[Discord] Alice in DM — " + ts,
		},
		{
			name:       "no channel access",
			authorName: "Charlie",
			want:       "[Discord] Charlie — " + ts,
		},
		{
			name: "no author or channel",
			want: "[Discord] Unknown — " + ts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildSubject(tt.channelName, tt.threadName, tt.isDM, tt.authorName, msgID)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSnowflakeTime(t *testing.T) {
	t.Parallel()

	got := snowflakeTime("1360000000000000000")
	want := "Apr 10, 9:14 PM"
	if got.Format("Jan 2, 3:04 PM") != want {
		t.Fatalf("got %q, want %q", got.Format("Jan 2, 3:04 PM"), want)
	}
}

func TestSnowflakeTime_invalid(t *testing.T) {
	t.Parallel()

	got := snowflakeTime("not-a-number")
	if !got.IsZero() {
		t.Fatalf("expected zero time for invalid ID, got %v", got)
	}
}
