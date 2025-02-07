package main

import (
	"os"

	command "github.com/gexec/gexec/pkg/command/server"
	"github.com/joho/godotenv"
)

func main() {
	if env := os.Getenv("GEXEC_ENV_FILE"); env != "" {
		_ = godotenv.Load(env)
	}

	if err := command.Run(); err != nil {
		os.Exit(1)
	}
}
