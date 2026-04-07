package cmd

import (
	"flag"
	"fmt"
	"os"
)

type options struct {
	version bool
	host    string
	port    string
	gateway bool

	discordToken     string
	discordAppID     string
	discordPublicKey string

	emailProvider string
	fromEmail     string
	toEmail       string

	// SMTP options
	smtpUser     string
	smtpPassword string

	// Resend options
	resendAPIKey string
}

func parseFlags(args []string) options {
	var opts options
	fs := flag.NewFlagSet("fwd2email", flag.ExitOnError)
	fs.BoolVar(&opts.version, "version", false, "Print version and exit")
	fs.StringVar(&opts.host, "host", envOrDefault("HOST", ""), "HTTP server host")
	fs.StringVar(&opts.port, "port", envOrDefault("PORT", "8080"), "HTTP server port")
	fs.BoolVar(&opts.gateway, "gateway", false, "Use gateway (websocket) mode instead of webhook HTTP server")
	fs.StringVar(&opts.discordToken, "discord-token", os.Getenv("DISCORD_TOKEN"), "Discord bot token")
	fs.StringVar(&opts.discordAppID, "discord-app-id", os.Getenv("DISCORD_APP_ID"), "Discord application ID")
	fs.StringVar(&opts.discordPublicKey, "discord-public-key", os.Getenv("DISCORD_PUBLIC_KEY"), "Discord public key for signature verification")
	fs.StringVar(&opts.emailProvider, "email-provider", envOrDefault("EMAIL_PROVIDER", "smtp"), "Email provider: smtp or resend")
	fs.StringVar(&opts.fromEmail, "from-email", os.Getenv("FROM_EMAIL"), "Sender email address")
	fs.StringVar(&opts.toEmail, "to-email", os.Getenv("TO_EMAIL"), "Recipient email address")
	fs.StringVar(&opts.smtpUser, "smtp-user", os.Getenv("SMTP_USER"), "SMTP username (smtp provider)")
	fs.StringVar(&opts.smtpPassword, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password (smtp provider)")
	fs.StringVar(&opts.resendAPIKey, "resend-api-key", os.Getenv("RESEND_API_KEY"), "Resend API key (resend provider)")
	_ = fs.Parse(args)
	return opts
}

func (o options) validate() error {
	if o.discordToken == "" {
		return fmt.Errorf("required config is not set: discord-token")
	}
	if o.discordAppID == "" {
		return fmt.Errorf("required config is not set: discord-app-id")
	}
	if !o.gateway && o.discordPublicKey == "" {
		return fmt.Errorf("required config is not set: discord-public-key")
	}
	if o.toEmail == "" {
		return fmt.Errorf("required config is not set: to-email")
	}
	switch o.emailProvider {
	case "smtp":
		if o.smtpUser == "" {
			return fmt.Errorf("required config is not set: smtp-user")
		}
		if o.smtpPassword == "" {
			return fmt.Errorf("required config is not set: smtp-password")
		}
	case "resend":
		if o.resendAPIKey == "" {
			return fmt.Errorf("required config is not set: resend-api-key")
		}
		if o.fromEmail == "" {
			return fmt.Errorf("required config is not set: from-email")
		}
	default:
		return fmt.Errorf("unknown email provider: %s (use smtp or resend)", o.emailProvider)
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
