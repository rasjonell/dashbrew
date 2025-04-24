package data

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type TodoOutput struct {
	Done  bool
	Title string
}

type FetchOutput interface {
	Error() error
	Output() string
}

type fetchOutput struct {
	err    error
	output string
}

func (f *fetchOutput) Output() string {
	return f.output
}

func (f *fetchOutput) Error() error {
	return f.err
}

func newFetchOutput(output string, err error) *fetchOutput {
	return &fetchOutput{
		err:    err,
		output: output,
	}
}

func RunScript(command string) FetchOutput {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return newFetchOutput("", fmt.Errorf("Empty command"))
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()
	return newFetchOutput(string(out), err)
}

func RunAPI(url string) FetchOutput {
	if url == "" {
		return newFetchOutput("", fmt.Errorf("Empty URL"))
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return newFetchOutput("", fmt.Errorf("HTTP GET Error: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return newFetchOutput("", fmt.Errorf("API Request Failed: status %d %s\n%s",
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			string(bodyBytes),
		))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return newFetchOutput("", fmt.Errorf("Failed to read response body: %w", err))
	}

	return newFetchOutput(string(bodyBytes), nil)
}

func ReadTodoFile(path string) ([]*TodoOutput, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var items []*TodoOutput
	for _, line := range strings.Split(string(bytes), "\n") {
		line = strings.TrimSpace(line)
		if len(line) < 2 {
			continue
		}

		done := line[0] == '+'
		title := strings.TrimSpace(line[1:])
		items = append(items, &TodoOutput{Title: title, Done: done})
	}

	return items, nil
}

func WriteTodoFile(path string, items []*TodoOutput) error {
	var lines []string
	for _, item := range items {
		prefix := "-"
		if item.Done {
			prefix = "+"
		}
		lines = append(lines, fmt.Sprintf("%s %s", prefix, item.Title))
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}
