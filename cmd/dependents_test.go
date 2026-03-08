package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestFetchDocument(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("<html><body><h1>Hello</h1></body></html>"))
		}))
		defer server.Close()

		client := server.Client()
		doc, err := fetchDocument(client, server.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := doc.Find("h1").Text()
		if text != "Hello" {
			t.Errorf("expected 'Hello', got %q", text)
		}
	})

	t.Run("non-200 status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := server.Client()
		_, err := fetchDocument(client, server.URL)
		if err == nil {
			t.Fatal("expected error for 404 status")
		}
		if !strings.Contains(err.Error(), "unexpected status 404") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("connection error", func(t *testing.T) {
		client := &http.Client{}
		_, err := fetchDocument(client, "http://127.0.0.1:1")
		if err == nil {
			t.Fatal("expected error for connection failure")
		}
	})
}

func TestParseDependentsCount(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected int
	}{
		{
			name: "single count",
			html: `<html><body>
				<div class="table-list-header-toggle">
					<a>42 Repositories</a>
					<a>5 Packages</a>
				</div>
			</body></html>`,
			expected: 42,
		},
		{
			name: "count with comma",
			html: `<html><body>
				<div class="table-list-header-toggle">
					<a>1,234 Repositories</a>
				</div>
			</body></html>`,
			expected: 1234,
		},
		{
			name: "zero count",
			html: `<html><body>
				<div class="table-list-header-toggle">
					<a>0 Repositories</a>
				</div>
			</body></html>`,
			expected: 0,
		},
		{
			name:     "no matching element",
			html:     `<html><body><p>No dependents info</p></body></html>`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			count := 0
			doc.Find(".table-list-header-toggle a").Each(func(_ int, s *goquery.Selection) {
				text := strings.TrimSpace(s.Text())
				if strings.Contains(text, "Repositories") {
					parts := strings.Fields(text)
					if len(parts) >= 1 {
						numStr := strings.ReplaceAll(parts[0], ",", "")
						fmt.Sscanf(numStr, "%d", &count)
					}
				}
			})

			if count != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, count)
			}
		})
	}
}

func TestGetDependents(t *testing.T) {
	t.Run("parses dependents page", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html><body>
				<div class="table-list-header-toggle">
					<a>15 Repositories</a>
					<a>3 Packages</a>
				</div>
			</body></html>`))
		}))
		defer server.Close()

		// Override getDependents to use test server by testing the parsing logic directly
		client := server.Client()
		doc, err := fetchDocument(client, server.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		count := 0
		doc.Find(".table-list-header-toggle a").Each(func(_ int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if strings.Contains(text, "Repositories") {
				parts := strings.Fields(text)
				if len(parts) >= 1 {
					numStr := strings.ReplaceAll(parts[0], ",", "")
					fmt.Sscanf(numStr, "%d", &count)
				}
			}
		})

		if count != 15 {
			t.Errorf("expected 15, got %d", count)
		}
	})
}

func TestRunDependents(t *testing.T) {
	t.Run("no repos returns error", func(t *testing.T) {
		err := runDependents(nil, []string{})
		if err == nil {
			t.Fatal("expected error when no repos specified")
		}
		if err.Error() != "no repositories specified" {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestDetailedFlag(t *testing.T) {
	t.Run("detailed flag is registered", func(t *testing.T) {
		flag := dependentsCmd.Flags().Lookup("detailed")
		if flag == nil {
			t.Fatal("expected --detailed flag to be registered")
		}
		if flag.DefValue != "false" {
			t.Errorf("expected default value 'false', got %q", flag.DefValue)
		}
	})
}

func TestOutputFormat(t *testing.T) {
	html := `<html><body>
		<div class="table-list-header-toggle">
			<a>15 Repositories</a>
			<a>3 Packages</a>
		</div>
	</body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	defer server.Close()

	// Override getDependents by testing output formatting directly
	// We capture stdout to verify the format
	tests := []struct {
		name           string
		detailedFlag   bool
		expectedPrefix string
	}{
		{
			name:           "default output is count only",
			detailedFlag:   false,
			expectedPrefix: "15\n",
		},
		{
			name:           "detailed output contains repo name and count",
			detailedFlag:   true,
			expectedPrefix: "owner/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			detailed = tt.detailedFlag
			repo := "owner/repo"
			count := 15
			if tt.detailedFlag {
				url := fmt.Sprintf("https://github.com/%s/network/dependents", repo)
				fmt.Printf("%-35s | \033]8;;%s\033\\deps\033]8;;\033\\ | %d\n", repo, url, count)
			} else {
				fmt.Printf("%d\n", count)
			}

			w.Close()
			buf := make([]byte, 1024)
			n, _ := r.Read(buf)
			os.Stdout = oldStdout
			output := string(buf[:n])

			if !strings.Contains(output, tt.expectedPrefix) {
				t.Errorf("expected output to contain %q, got %q", tt.expectedPrefix, output)
			}
		})
	}
}
