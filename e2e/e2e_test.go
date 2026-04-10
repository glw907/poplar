package e2e

import (
	"bytes"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	binary       string
	updateGolden = flag.Bool("update-golden", false, "regenerate golden files")
)

func TestMain(m *testing.M) {
	flag.Parse()

	// Build the binary once.
	tmp, err := os.MkdirTemp("", "mailrender-test")
	if err != nil {
		panic(err)
	}
	binary = filepath.Join(tmp, "mailrender")
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/mailrender")
	cmd.Dir = filepath.Join("..")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("build failed: " + err.Error())
	}

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

func TestHTMLFixtures(t *testing.T) {
	fixtures, err := filepath.Glob("testdata/*.html")
	if err != nil {
		t.Fatalf("globbing fixtures: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no HTML fixtures found in testdata/")
	}

	for _, fixture := range fixtures {
		name := strings.TrimSuffix(filepath.Base(fixture), ".html")
		t.Run(name, func(t *testing.T) {
			input, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatalf("reading fixture: %v", err)
			}

			cmd := exec.Command(binary, "html")
			cmd.Stdin = bytes.NewReader(input)
			cmd.Env = append(os.Environ(), "AERC_COLUMNS=80")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("running html filter: %v\noutput: %s", err, out)
			}

			goldenPath := filepath.Join("testdata", "golden", name+".txt")
			if *updateGolden {
				os.MkdirAll(filepath.Dir(goldenPath), 0755)
				if err := os.WriteFile(goldenPath, out, 0644); err != nil {
					t.Fatalf("writing golden file: %v", err)
				}
				return
			}

			golden, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("reading golden file (run with -update-golden to create): %v", err)
			}
			if !bytes.Equal(out, golden) {
				t.Errorf("output differs from golden file %s\ngot:\n%s\nwant:\n%s", goldenPath, out, golden)
			}
		})
	}
}

func TestHeadersFixture(t *testing.T) {
	input := "From: Alice <alice@example.com>\r\nTo: Bob <bob@example.com>, Charlie <charlie@example.com>\r\nDate: Mon, 01 Jan 2026 00:00:00 +0000\r\nSubject: Test Message\r\nX-Mailer: test\r\n\r\n"

	cmd := exec.Command(binary, "headers")
	cmd.Stdin = strings.NewReader(input)
	cmd.Env = append(os.Environ(), "AERC_COLUMNS=80")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running headers filter: %v\noutput: %s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "From:") {
		t.Error("missing From header")
	}
	if !strings.Contains(output, "Subject:") {
		t.Error("missing Subject header")
	}
	if strings.Contains(output, "X-Mailer") {
		t.Error("X-Mailer should be stripped")
	}
}
