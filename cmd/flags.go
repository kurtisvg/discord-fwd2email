package cmd

import (
	"flag"
	"os"
)

type options struct {
	port string
}

func parseFlags(args []string) options {
	var opts options
	fs := flag.NewFlagSet("discord-forward-to-email", flag.ExitOnError)
	fs.StringVar(&opts.port, "port", envOrDefault("PORT", "8080"), "HTTP server port")
	_ = fs.Parse(args)
	return opts
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
