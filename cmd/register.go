package cmd

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

func runRegister() {
	token := requireEnv("DISCORD_TOKEN")
	appID := requireEnv("DISCORD_APP_ID")

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		slog.Error("failed to create discord session", "error", err)
		return
	}

	cmd, err := s.ApplicationCommandCreate(appID, "", &discordgo.ApplicationCommand{
		Name: "Forward to inbox",
		Type: discordgo.MessageApplicationCommand,
	})
	if err != nil {
		slog.Error("failed to register command", "error", err)
		return
	}

	slog.Info("registered command", "name", cmd.Name, "id", cmd.ID)
}
