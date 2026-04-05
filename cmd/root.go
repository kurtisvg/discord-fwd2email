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
	if err := opts.validate(); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mailer := email.NewMailer(opts.gmailUser, opts.gmailAppPassword)
	handler, err := discord.NewHandler(opts.discordPublicKey, opts.discordToken, opts.discordAppID, opts.gmailUser, mailer)
	if err != nil {
		slog.Error("failed to create handler", "error", err)
		os.Exit(1)
	}

	if err := handler.RegisterCommand(); err != nil {
		slog.Error("failed to register command", "error", err)
		os.Exit(1)
	}
	slog.Info("registered command")

	if opts.gateway {
		runGateway(ctx, handler)
	} else {
		runWebhook(ctx, handler, opts.port)
	}
}

func runWebhook(ctx context.Context, handler *discord.Handler, port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/interactions", handler.HandleInteraction)

	srv := &http.Server{Addr: ":" + port, Handler: mux}

	go func() {
		<-ctx.Done()
		slog.Info("shutting down")
		_ = srv.Shutdown(context.Background())
	}()

	slog.Info("listening (webhook mode)", "port", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func runGateway(ctx context.Context, handler *discord.Handler) {
	s := handler.Session()
	s.AddHandler(handler.HandleGatewayInteraction)

	if err := s.Open(); err != nil {
		slog.Error("failed to open gateway connection", "error", err)
		os.Exit(1)
	}
	defer func() { _ = s.Close() }()

	slog.Info("connected (gateway mode)")
	<-ctx.Done()
	slog.Info("shutting down")
}
