package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rasjonell/dashbrew/internal/config"
	"github.com/rasjonell/dashbrew/internal/tui"
)

func main() {
	cfg, err := config.LoadConfig("dashboard.json")
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

	clear()
}

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	}

	cmd.Run()
}
