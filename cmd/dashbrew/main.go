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

var (
	configPath string
	themeFlag  string
	listThemes bool
)

func main() {
	flag.StringVar(&configPath, "c", "dashboard.json", "Path to dashboard config.")
	flag.StringVar(&themeFlag, "t", "", "Override style.theme from the config (e.g. dracula, nord).")
	flag.StringVar(&themeFlag, "theme", "", "Override style.theme from the config (e.g. dracula, nord).")
	flag.BoolVar(&listThemes, "list-themes", false, "List available themes and exit.")
	flag.Parse()

	if listThemes {
		for _, name := range config.AvailableThemes() {
			fmt.Println(name)
		}
		return
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if themeFlag != "" {
		if err := config.OverrideTheme(cfg, themeFlag); err != nil {
			fmt.Printf("Failed to apply theme override: %v\n", err)
			os.Exit(1)
		}
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
