package cmd

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func Execute() {
	register := flag.Bool("register", false, "register the Discord message command and exit")
	port := flag.String("port", envOrDefault("PORT", "8080"), "HTTP server port")
	flag.Parse()

	if *register {
		runRegister()
		return
	}

	_ = port // will be used in M1.2
	fmt.Println("server not yet implemented")
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("required environment variable is not set", "key", key)
		os.Exit(1)
	}
	return v
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
