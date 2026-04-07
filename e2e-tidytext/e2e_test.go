package e2e_tidytext

import (
	"bytes"
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binary string

func TestMain(m *testing.M) {
	flag.Parse()
	tmp, err := os.MkdirTemp("", "tidytext-test")
	if err != nil {
		panic(err)
	}
	binary = filepath.Join(tmp, "tidytext")
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/tidytext")
	cmd.Dir = filepath.Join("..")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("build failed: " + err.Error())
	}
	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

// mockAPIServer returns a test HTTP server that responds with a fixed corrected string.
func mockAPIServer(corrected string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		}{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{{Type: "text", Text: corrected}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestE2E_NoArgs_ShowsHelp(t *testing.T) {
	cmd := exec.Command(binary)
	out, _ := cmd.CombinedOutput()
	output := string(out)
	if !strings.Contains(output, "fix") {
		t.Errorf("expected help output to contain 'fix', got:\n%s", output)
	}
	if !strings.Contains(output, "config") {
		t.Errorf("expected help output to contain 'config', got:\n%s", output)
	}
}

func TestE2E_FixNoInput_ShowsError(t *testing.T) {
	// When stdin is /dev/null (not a TTY, but empty), the binary reads zero
	// bytes, so Tidy sees empty input and reports "no author text found".
	// The "no input provided" message only appears when stdin is a real TTY
	// (interactive use), which cannot be simulated in subprocess tests.
	devnull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open /dev/null: %v", err)
	}
	defer devnull.Close()

	cmd := exec.Command(binary, "fix", "--no-config")
	cmd.Stdin = devnull
	cmd.Env = append(os.Environ(), "ANTHROPIC_API_KEY=test-key")
	out, _ := cmd.CombinedOutput()
	output := string(out)
	// Empty input produces "no author text found" (safe fallback, exit 0).
	if !strings.Contains(output, "no author text found") {
		t.Errorf("expected 'no author text found' message, got:\n%s", output)
	}
}

func TestE2E_FixPipe(t *testing.T) {
	corrected := "This is corrected text.\n"
	server := mockAPIServer(corrected)
	defer server.Close()

	input := "This is some text with erors.\n"
	cmd := exec.Command(binary, "fix", "--no-config")
	cmd.Stdin = bytes.NewBufferString(input)
	cmd.Env = append(os.Environ(),
		"ANTHROPIC_API_KEY=test-key",
		"TIDYTEXT_API_URL="+server.URL,
	)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("tidytext fix: %v", err)
	}
	if !strings.Contains(string(out), "corrected text") {
		t.Errorf("expected corrected text in stdout, got:\n%s", string(out))
	}
}

func TestE2E_FixFile(t *testing.T) {
	corrected := "This is corrected text.\n"
	server := mockAPIServer(corrected)
	defer server.Close()

	original := "This is some text with erors.\n"
	f, err := os.CreateTemp("", "tidytext-input-*.txt")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(original)
	f.Close()

	cmd := exec.Command(binary, "fix", "--no-config", f.Name())
	cmd.Env = append(os.Environ(),
		"ANTHROPIC_API_KEY=test-key",
		"TIDYTEXT_API_URL="+server.URL,
	)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("tidytext fix file: %v", err)
	}
	if !strings.Contains(string(out), "corrected text") {
		t.Errorf("expected corrected text in stdout, got:\n%s", string(out))
	}

	// Verify original file is unchanged.
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("read original file: %v", err)
	}
	if string(data) != original {
		t.Errorf("original file was modified: got %q, want %q", string(data), original)
	}
}

func TestE2E_FixFileInPlace(t *testing.T) {
	corrected := "This is corrected text.\n"
	server := mockAPIServer(corrected)
	defer server.Close()

	original := "This is some text with erors.\n"
	f, err := os.CreateTemp("", "tidytext-inplace-*.txt")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(original)
	f.Close()

	cmd := exec.Command(binary, "fix", "--no-config", "--in-place", f.Name())
	cmd.Env = append(os.Environ(),
		"ANTHROPIC_API_KEY=test-key",
		"TIDYTEXT_API_URL="+server.URL,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("tidytext fix --in-place: %v", err)
	}

	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("read in-place file: %v", err)
	}
	if !strings.Contains(string(data), "corrected text") {
		t.Errorf("expected file to contain corrected text, got:\n%s", string(data))
	}
}

func TestE2E_PreservesQuotedText(t *testing.T) {
	// The mock returns only corrected author lines; quoted lines must be preserved.
	correctedAuthor := "This is the corrected reply.\n"
	server := mockAPIServer(correctedAuthor)
	defer server.Close()

	input := "This is the reply with erors.\n> This is the quoted original.\n"
	cmd := exec.Command(binary, "fix", "--no-config")
	cmd.Stdin = bytes.NewBufferString(input)
	cmd.Env = append(os.Environ(),
		"ANTHROPIC_API_KEY=test-key",
		"TIDYTEXT_API_URL="+server.URL,
	)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("tidytext fix: %v", err)
	}
	output := string(out)
	if !strings.Contains(output, "> This is the quoted original.") {
		t.Errorf("expected quoted line preserved in output, got:\n%s", output)
	}
	if !strings.Contains(output, "corrected reply") {
		t.Errorf("expected corrected author text in output, got:\n%s", output)
	}
}

func TestE2E_MissingAPIKey(t *testing.T) {
	input := "This is some text.\n"
	cmd := exec.Command(binary, "fix", "--no-config")
	cmd.Stdin = bytes.NewBufferString(input)
	// Strip ANTHROPIC_API_KEY from environment.
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("tidytext fix (no api key): %v", err)
	}
	// Should return original text unchanged (safe fallback).
	if !strings.Contains(string(out), "This is some text.") {
		t.Errorf("expected original text in output (safe fallback), got:\n%s", string(out))
	}
}

func TestE2E_ConfigShow(t *testing.T) {
	cmd := exec.Command(binary, "config")
	// Point HOME to a temp dir so config init can't create a real config.
	tmp := t.TempDir()
	cmd.Env = append(os.Environ(), "HOME="+tmp)
	out, _ := cmd.CombinedOutput()
	output := string(out)
	if !strings.Contains(output, "spelling") {
		t.Errorf("expected 'spelling' in config output, got:\n%s", output)
	}
}

func TestE2E_Version(t *testing.T) {
	cmd := exec.Command(binary, "--version")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("tidytext --version: %v", err)
	}
	if !strings.Contains(string(out), "tidytext") {
		t.Errorf("expected 'tidytext' in version output, got:\n%s", string(out))
	}
}
