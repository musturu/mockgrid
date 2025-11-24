package main

import (
	"log/slog"
	"os"

	"github.com/mustur/mockgrid/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("execution failed", "error", err)
		os.Exit(1)
	}
}
