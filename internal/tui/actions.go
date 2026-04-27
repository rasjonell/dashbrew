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

// actionStartMsg flips a focused component into loading state at the moment
// the user fires an action (key chord or palette pick). The runActionCmd is
// dispatched in the same Batch so the spinner and the shell process start
// together.
type actionStartMsg struct {
	ComponentID string
	Label       string
	OutputAs    string
	Type        string
	PendingID   string
}

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

// runActionCmd builds the tea.Cmd that executes an action asynchronously.
// fire_and_forget returns nil immediately; the process is reaped silently in
// a goroutine. replace_pane captures stdout/stderr and emits actionResultMsg
// or actionFailedMsg.
func runActionCmd(componentID, pendingID, label string, a config.ActionConfig) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var cancel context.CancelFunc
		switch {
		case a.TimeoutSeconds > 0:
			ctx, cancel = context.WithTimeout(ctx, time.Duration(a.TimeoutSeconds)*time.Second)
			defer cancel()
		case a.TimeoutSeconds == 0 && a.Type == "fire_and_forget":
			// fire_and_forget gets the standard guard so a runaway script doesn't leak.
			ctx, cancel = context.WithTimeout(ctx, defaultActionTimeout)
			defer cancel()
		case a.TimeoutSeconds == 0 && a.Type == "replace_pane":
			ctx, cancel = context.WithTimeout(ctx, defaultActionTimeout)
			defer cancel()
		}
		// TimeoutSeconds < 0 is rejected at validation; explicit 0 with a "no-timeout"
		// opt-in is reserved for v3 once we support streaming.

		cmd := exec.CommandContext(ctx, "sh", "-c", a.Command)

		if a.Type == "fire_and_forget" {
			cmd.Stdout = io.Discard
			cmd.Stderr = io.Discard
			if err := cmd.Start(); err != nil {
				return actionFailedMsg{
					ComponentID: componentID, PendingID: pendingID,
					Command: a.Command, ExitCode: -1, Err: err, Label: label,
				}
			}
			go func() {
				_ = cmd.Wait()
				if cancel != nil {
					cancel()
				}
			}()
			return nil
		}

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
}
