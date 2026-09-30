package knowledge

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Build combines the docs and the block catalog into one markdown document.
// It's small enough to give the model all of it with every question, so there
// is nothing to search.
func Build(docsPath, catalogPath, docsBaseURL string) (string, error) {
	raw, err := os.ReadFile(catalogPath)
	if err != nil {
		return "", fmt.Errorf("read catalog: %w", err)
	}
	catalog, err := ParseCatalog(raw)
	if err != nil {
		return "", err
	}

	docs, err := LoadDocs(docsPath, docsBaseURL, catalog)
	if err != nil {
		return "", fmt.Errorf("load docs: %w", err)
	}
	if len(docs) == 0 {
		return "", fmt.Errorf("no docs found in %s", docsPath)
	}

	var b strings.Builder
	described := map[string]bool{}
	for _, d := range docs {
		fmt.Fprintf(&b, "=== Page: %s\n%s%s\n\n%s\n\n", d.Title, urlPrefix, d.URL, d.Body)
		for _, t := range d.Blocks {
			described[t] = true
		}
	}

	// Blocks that are added together with another one, like the branches of a
	// condition, have no page of their own.
	var rest []string
	for t := range catalog.Nodes {
		if !described[t] {
			rest = append(rest, t)
		}
	}
	sort.Strings(rest)
	if len(rest) > 0 {
		b.WriteString("=== Blocks without their own page\n\n")
		for _, t := range rest {
			block, _ := catalog.Describe(t)
			b.WriteString(block)
			b.WriteString("\n\n")
		}
	}

	return strings.TrimSpace(b.String()) + "\n", nil
}

// urlPrefix starts the line with the URL of each page.
const urlPrefix = "URL: "

// PageURLs returns the URLs of the pages in the knowledge, the only ones the
// bot links to.
func PageURLs(knowledge string) []string {
	var urls []string
	for _, line := range strings.Split(knowledge, "\n") {
		if url, ok := strings.CutPrefix(line, urlPrefix); ok {
			urls = append(urls, url)
		}
	}
	return urls
}
