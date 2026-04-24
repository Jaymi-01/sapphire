package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// 1. Setup Logging
	logFile, err := os.OpenFile("game.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	handler := slog.NewTextHandler(logFile, nil)
	slog.SetDefault(slog.New(handler))

	slog.Info("Starting Sapphire RPG...")

	// 2. Initialize Database
	if err := InitDB("althoria.db"); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		fmt.Printf("Database Error: %v\nPress Enter to exit...", err)
		fmt.Scanln()
		os.Exit(1)
	}
	defer CloseDB()

	// 3. Setup Model and Load Data
	m := NewModel()
	if err := m.LoadGameState(); err != nil {
		slog.Error("Failed to load game state", "error", err)
		fmt.Printf("State Error: %v\nPress Enter to exit...", err)
		fmt.Scanln()
		os.Exit(1)
	}

	// 4. Handle Graceful Shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		slog.Info("Shutdown signal received. Exiting...")
		CloseDB()
		os.Exit(0)
	}()

	// 5. Run Bubble Tea Program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		slog.Error("Error running program", "error", err)
		fmt.Printf("Launch Error: %v\nPress Enter to exit...", err)
		fmt.Scanln()
		os.Exit(1)
	}
}
