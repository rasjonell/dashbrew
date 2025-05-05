package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/oliveagle/jsonpath"
)

type FetchOutput interface {
	Error() error
	Output() string
}

type fetchOutput struct {
	err    error
	output string
}

type TodoOutput struct {
	Done  bool
	Title string
}

func (f *fetchOutput) Error() error   { return f.err }
func (f *fetchOutput) Output() string { return f.output }

func NewFetchOutput(output string, err error) *fetchOutput {
	return &fetchOutput{
		err:    err,
		output: output,
	}
}

func RunScript(command string) FetchOutput {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return NewFetchOutput("", fmt.Errorf("Empty command"))
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()
	return NewFetchOutput(string(out), err)
}

func RunAPI(url, jsonPath string) FetchOutput {
	if url == "" {
		return NewFetchOutput("", fmt.Errorf("Empty URL"))
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return NewFetchOutput("", fmt.Errorf("HTTP GET Error: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return NewFetchOutput("", fmt.Errorf("API Request Failed: status %d %s\n%s",
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			string(bodyBytes),
		))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewFetchOutput("", fmt.Errorf("Failed to read response body: %w", err))
	}

	if jsonPath == "" {
		return NewFetchOutput(string(bodyBytes), nil)
	}

	var jsonData any
	err = json.Unmarshal(bodyBytes, &jsonData)
	if err != nil {
		return NewFetchOutput("", fmt.Errorf("Failed to parse API response: %w", err))
	}

	res, err := jsonpath.JsonPathLookup(jsonData, jsonPath)
	if err != nil {
		return NewFetchOutput("", fmt.Errorf("Failed to lookup json path '%s': %w", jsonPath, err))
	}

	var resultBytes []byte
	var marshallErr error

	if res == nil {
		resultBytes = []byte("null")
	} else {
		resultBytes, marshallErr = json.MarshalIndent(res, "", "  ")
	}

	if marshallErr != nil {
		return NewFetchOutput("", fmt.Errorf("Failed to marshal jsonPath result: %w", marshallErr))
	}

	return NewFetchOutput(string(resultBytes), nil)
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
