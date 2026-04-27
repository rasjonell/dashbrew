package tui

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rasjonell/dashbrew/internal/config"
)

const defaultActionTimeout = 30 * time.Second

// actionResultMsg is emitted when a replace_pane action exits successfully
// (zero exit, within timeout). fire_and_forget actions never emit this.
type actionResultMsg struct {
	ComponentID string
	PendingID   string
	Stdout      string
	OutputAs    string
	Label       string
}

// actionFailedMsg is emitted on non-zero exit, spawn failure, or timeout.
type actionFailedMsg struct {
	ComponentID string
	PendingID   string
	Command     string
	ExitCode    int
	Stderr      string
	Err         error
	Label       string
}

var actionCounter atomic.Uint64

// nextPendingID mints a per-action token used to discard late-arriving results
// after the user has reset (Esc/R) or fired a different action on the same
// component.
func nextPendingID() string {
	return fmt.Sprintf("act-%d", actionCounter.Add(1))
}

// matchBinding returns the binding whose Key (after normalization) matches the
// pressed chord, or nil if none does.
func matchBinding(bindings []config.Binding, chord string) *config.Binding {
	chord = config.NormalizeKey(chord)
	for i, b := range bindings {
		if config.NormalizeKey(b.Key) == chord {
			return &bindings[i]
		}
	}
	return nil
}

// dispatchAction starts an action on a component: synchronously transitions
// the component into loading state and returns the tea.Cmd that runs the
// shell command and emits the result/failed message.
func (m *model) dispatchAction(componentID, label string, a config.ActionConfig) tea.Cmd {
	comp, ok := m.components[componentID]
	if !ok {
		return nil
	}
	pendingID := nextPendingID()
	updated, beginCmd := comp.BeginAction(label, a.OutputAs, pendingID)
	m.components[componentID] = updated
	runCmd := runActionCmd(componentID, pendingID, label, a)
	return tea.Batch(beginCmd, runCmd)
}

// runActionCmd builds the tea.Cmd that executes an action asynchronously.
// fire_and_forget returns nil immediately; the process is reaped silently in
// a goroutine that owns its own context cancel so the spawned process isn't
// killed when the closure returns. replace_pane captures stdout/stderr and
// emits actionResultMsg or actionFailedMsg.
func runActionCmd(componentID, pendingID, label string, a config.ActionConfig) tea.Cmd {
	return func() tea.Msg {
		timeout := time.Duration(a.TimeoutSeconds) * time.Second
		if a.TimeoutSeconds <= 0 {
			timeout = defaultActionTimeout
		}

		if a.Type == "fire_and_forget" {
			return runFireAndForget(componentID, pendingID, label, a, timeout)
		}
		return runReplacePane(componentID, pendingID, label, a, timeout)
	}
}

func runFireAndForget(componentID, pendingID, label string, a config.ActionConfig, timeout time.Duration) tea.Msg {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	cmd := exec.CommandContext(ctx, "sh", "-c", a.Command)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		cancel()
		return actionFailedMsg{
			ComponentID: componentID, PendingID: pendingID,
			Command: a.Command, ExitCode: -1, Err: err, Label: label,
		}
	}

	// Goroutine owns the cancel so the spawned process survives until it
	// either exits on its own or hits the timeout. Returning nil to bubbletea
	// here MUST NOT kill the process.
	go func() {
		defer cancel()
		_ = cmd.Wait()
	}()
	return nil
}

func runReplacePane(componentID, pendingID, label string, a config.ActionConfig, timeout time.Duration) tea.Msg {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", a.Command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return actionFailedMsg{
			ComponentID: componentID, PendingID: pendingID,
			Command: a.Command, ExitCode: -1, Stderr: stderr.String(),
			Err: ctx.Err(), Label: label,
		}
	}
	if err != nil {
		exitCode := -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		return actionFailedMsg{
			ComponentID: componentID, PendingID: pendingID,
			Command: a.Command, ExitCode: exitCode, Stderr: stderr.String(),
			Err: err, Label: label,
		}
	}

	return actionResultMsg{
		ComponentID: componentID, PendingID: pendingID,
		Stdout: stdout.String(), OutputAs: a.OutputAs, Label: label,
	}
}
