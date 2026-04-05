package cmd

import (
	"log/slog"
	"os"

	"github.com/bwmarrin/discordgo"
)

func runRegister(opts options) {
	s, err := discordgo.New("Bot " + opts.discordToken)
	if err != nil {
		slog.Error("failed to create discord session", "error", err)
		os.Exit(1)
	}

	cmd, err := s.ApplicationCommandCreate(opts.discordAppID, "", &discordgo.ApplicationCommand{
		Name: "Forward to inbox",
		Type: discordgo.MessageApplicationCommand,
	})
	if err != nil {
		slog.Error("failed to register command", "error", err)
		os.Exit(1)
	}

	slog.Info("registered command", "name", cmd.Name, "id", cmd.ID)
}
