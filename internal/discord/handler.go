package discord

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/discord-forward-to-email/internal/email"
)

type Handler struct {
	publicKey ed25519.PublicKey
	session   *discordgo.Session
	appID     string
	gmailUser string
	mailer    *email.Mailer
}

func NewHandler(publicKeyHex, token, appID, gmailUser string, mailer *email.Mailer) (*Handler, error) {
	key, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid discord public key: %w", err)
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}

	return &Handler{
		publicKey: ed25519.PublicKey(key),
		session:   session,
		appID:     appID,
		gmailUser: gmailUser,
		mailer:    mailer,
	}, nil
}

func (h *Handler) HandleInteraction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if !h.verifySignature(r, body) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var interaction discordgo.Interaction
	if err := json.Unmarshal(body, &interaction); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	switch interaction.Type {
	case discordgo.InteractionPing:
		respondJSON(w, discordgo.InteractionResponse{
			Type: discordgo.InteractionResponsePong,
		})

	case discordgo.InteractionApplicationCommand:
		respondJSON(w, discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		go h.handleForward(&interaction)

	default:
		http.Error(w, "unknown interaction type", http.StatusBadRequest)
	}
}

func (h *Handler) handleForward(interaction *discordgo.Interaction) {
	data := interaction.ApplicationCommandData()

	if len(data.Resolved.Messages) == 0 {
		h.editReply(interaction, "❌ Failed to forward — no message data in interaction.")
		return
	}

	var targetMsg *discordgo.Message
	for _, msg := range data.Resolved.Messages {
		targetMsg = msg
		break
	}

	guildID := interaction.GuildID
	if guildID == "" {
		guildID = "@me"
	}
	messageLink := fmt.Sprintf("https://discord.com/channels/%s/%s/%s",
		guildID, interaction.ChannelID, targetMsg.ID)

	// Fetch context messages (up to 5 before the target).
	contextMessages, contextErr := fetchContext(h.session, interaction.ChannelID, targetMsg.ID)
	if contextErr != nil {
		slog.Info("context fetch failed, forwarding target only", "error", contextErr)
	}

	channelName := ""
	channel, err := h.session.Channel(interaction.ChannelID)
	if err == nil {
		channelName = channel.Name
	}

	serverName := ""
	if interaction.GuildID != "" {
		guild, err := h.session.Guild(interaction.GuildID)
		if err == nil {
			serverName = guild.Name
		}
	}

	target := messageData(targetMsg)

	emailData := email.ForwardData{
		ServerName:      serverName,
		ChannelName:     channelName,
		MessageLink:     messageLink,
		ContextMessages: contextMessages,
		TargetMessage:   target,
	}

	subject := "[Discord] Forwarded chat"
	if channelName != "" {
		subject = fmt.Sprintf("[Discord] Forwarded chat in #%s", channelName)
	} else if serverName == "" && target.AuthorName != "" {
		subject = fmt.Sprintf("[Discord] Forwarded DM with %s", target.AuthorName)
	}

	if err := h.mailer.Send(h.gmailUser, subject, emailData); err != nil {
		slog.Error("email send failed", "error", err)
		h.editReply(interaction, "❌ Failed to forward — check bot logs.")
		return
	}

	if contextErr != nil {
		h.editReply(interaction, fmt.Sprintf("✉️ Forwarded to %s (target message only — no channel access for context)", h.gmailUser))
	} else {
		h.editReply(interaction, fmt.Sprintf("✉️ Forwarded to %s (with %d messages of context)", h.gmailUser, len(contextMessages)))
	}
}

func fetchContext(s *discordgo.Session, channelID, beforeMessageID string) ([]email.MessageData, error) {
	messages, err := s.ChannelMessages(channelID, 5, beforeMessageID, "", "")
	if err != nil {
		return nil, err
	}

	// ChannelMessages returns newest-first; reverse to oldest-first.
	context := make([]email.MessageData, len(messages))
	for i, msg := range messages {
		context[len(messages)-1-i] = messageData(msg)
	}
	return context, nil
}

func messageData(msg *discordgo.Message) email.MessageData {
	authorName := msg.Author.GlobalName
	if authorName == "" {
		authorName = msg.Author.Username
	}

	var attachments []email.Attachment
	for _, a := range msg.Attachments {
		attachments = append(attachments, email.Attachment{
			Filename: a.Filename,
			URL:      a.URL,
			IsImage:  strings.HasPrefix(a.ContentType, "image/"),
		})
	}

	return email.MessageData{
		AuthorName:  authorName,
		AvatarURL:   avatarURL(msg.Author),
		Content:     msg.Content,
		Attachments: attachments,
	}
}

func avatarURL(u *discordgo.User) string {
	if u.Avatar != "" {
		return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png?size=64", u.ID, u.Avatar)
	}
	// Default avatar index: (user_id >> 22) % 6
	// For simplicity, use the discriminator-based fallback which discordgo provides.
	return u.AvatarURL("64")
}

func (h *Handler) editReply(interaction *discordgo.Interaction, content string) {
	_, err := h.session.InteractionResponseEdit(interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		slog.Error("failed to edit interaction reply", "error", err)
	}
}

func (h *Handler) verifySignature(r *http.Request, body []byte) bool {
	sig, err := hex.DecodeString(r.Header.Get("X-Signature-Ed25519"))
	if err != nil {
		return false
	}
	timestamp := r.Header.Get("X-Signature-Timestamp")
	if timestamp == "" {
		return false
	}
	msg := append([]byte(timestamp), body...)
	return ed25519.Verify(h.publicKey, msg, sig)
}

func respondJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
