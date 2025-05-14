package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rasjonell/dashbrew/internal/config"
	"github.com/rasjonell/dashbrew/internal/tui"
)

var configPath string

func main() {
	flag.StringVar(&configPath, "c", "dashboard.json", "Path to dashboard config.")
	flag.Parse()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(tui.New(cfg), tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	if err != nil {
		fmt.Printf("Failed to start program: %v", err)
		os.Exit(1)
	}
}

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	}

	cmd.Run()
}
