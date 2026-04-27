package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const errorModalStderrLines = 12

type errorModalState struct {
	Command  string
	ExitCode int
	Stderr   string
	Err      error
	Label    string
}

func (m *model) renderErrorModal() string {
	s := m.errorModal
	if s == nil {
		return ""
	}

	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff5555")).Render("Action failed: " + s.Label)

	body := []string{header, ""}
	body = append(body, fmt.Sprintf("Command: %s", s.Command))
	if s.ExitCode != 0 || s.Err == nil {
		body = append(body, fmt.Sprintf("Exit:    %d", s.ExitCode))
	}
	if s.Err != nil {
		body = append(body, fmt.Sprintf("Error:   %s", s.Err.Error()))
	}
	if trimmed := strings.TrimSpace(s.Stderr); trimmed != "" {
		body = append(body, "", "stderr:")
		body = append(body, lastLines(trimmed, errorModalStderrLines))
	}
	body = append(body, "", lipgloss.NewStyle().Faint(true).Render("[Esc] dismiss"))

	width := min(max(m.width*60/100, 30), m.width-4)

	content := strings.Join(body, "\n")
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#ff5555")).
		Padding(1, 2).
		Width(width).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func lastLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}
