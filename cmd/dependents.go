package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

var detailed bool

var dependentsCmd = &cobra.Command{
	Use:   "dependents [owner/repo ...]",
	Short: "Get GitHub dependents count for repositories",
	Long:  "Fetches the number of repository dependents from GitHub's dependency graph.",
	RunE:  runDependents,
}

func init() {
	dependentsCmd.Flags().BoolVar(&detailed, "detailed", false, "Show detailed output with repo name, link, and count")
}

func runDependents(cmd *cobra.Command, args []string) error {
	repos := args

	if len(repos) == 0 {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					repos = append(repos, line)
				}
			}
		}
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repositories specified")
	}

	for _, repo := range repos {
		count, err := getDependents(repo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching dependents for %s: %v\n", repo, err)
			continue
		}
		if detailed {
			url := fmt.Sprintf("https://github.com/%s/network/dependents", repo)
			fmt.Printf("%-35s | \033]8;;%s\033\\deps\033]8;;\033\\ | %d\n", repo, url, count)
		} else {
			fmt.Printf("%d\n", count)
		}
	}

	return nil
}

func getDependents(repo string) (int, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://github.com/%s/network/dependents", repo)

	doc, err := fetchDocument(client, url)
	if err != nil {
		return 0, err
	}

	// Check for multiple packages by looking for package_id in links
	var packageURL string
	doc.Find("a[href*='package_id=']").Each(func(_ int, s *goquery.Selection) {
		if packageURL != "" {
			return
		}
		href, exists := s.Attr("href")
		if !exists {
			return
		}
		// Find the label matching the repo name
		label := strings.TrimSpace(s.Find(".select-menu-item-text").Text())
		if label == "" {
			// The label might be a sibling or the link text itself
			label = strings.TrimSpace(s.Text())
		}
		if label == repo {
			packageURL = "https://github.com" + href
		}
	})

	if packageURL != "" {
		doc, err = fetchDocument(client, packageURL)
		if err != nil {
			return 0, err
		}
	}

	// Extract the dependents count from the box-header
	count := 0
	doc.Find(".table-list-header-toggle a").Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if strings.Contains(text, "Repositories") {
			// Parse the number from text like "42 Repositories"
			parts := strings.Fields(text)
			if len(parts) >= 1 {
				numStr := strings.ReplaceAll(parts[0], ",", "")
				fmt.Sscanf(numStr, "%d", &count)
			}
		}
	})

	return count, nil
}

func fetchDocument(client *http.Client, url string) (*goquery.Document, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML from %s: %w", url, err)
	}

	return doc, nil
}
