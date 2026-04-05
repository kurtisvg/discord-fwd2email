package cmd

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/discord-forward-to-email/internal/discord"
	"github.com/discord-forward-to-email/internal/email"
)

func Execute() {
	opts := parseFlags(os.Args[1:])
	opts.validate()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mailer := email.NewMailer(opts.gmailUser, opts.gmailAppPassword)
	handler := discord.NewHandler(opts.discordPublicKey, opts.discordToken, opts.discordAppID, opts.gmailUser, mailer)

	mux := http.NewServeMux()
	mux.HandleFunc("/interactions", handler.HandleInteraction)

	srv := &http.Server{Addr: ":" + opts.port, Handler: mux}

	go func() {
		<-ctx.Done()
		slog.Info("shutting down")
		_ = srv.Close()
	}()

	slog.Info("listening", "port", opts.port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
