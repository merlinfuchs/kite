package index

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Document struct {
	Source string
	Title  string
	URL    string
	Body   string
}

func WalkDocs(root string, baseURL string) ([]Document, error) {
	var docs []Document

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") && !strings.HasSuffix(path, ".mdx") {
			return nil
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		body, fm := splitFrontmatter(string(raw))
		title := fm["title"]
		if title == "" {
			title = firstHeading(body)
		}
		if title == "" {
			title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		}

		rel, _ := filepath.Rel(root, path)
		slug := slugFromPath(rel, fm["slug"])

		docs = append(docs, Document{
			Source: rel,
			Title:  title,
			URL:    joinURL(baseURL, slug),
			Body:   body,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func WalkConcepts(path string) ([]Document, error) {
	if path == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return []Document{{
		Source: filepath.Base(path),
		Title:  "Kite Concepts",
		URL:    "",
		Body:   string(raw),
	}}, nil
}

func splitFrontmatter(s string) (string, map[string]string) {
	fm := map[string]string{}
	if !strings.HasPrefix(s, "---") {
		return s, fm
	}
	end := strings.Index(s[3:], "---")
	if end < 0 {
		return s, fm
	}
	header := s[3 : 3+end]
	body := strings.TrimLeft(s[3+end+3:], "\n")
	for _, line := range strings.Split(header, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		fm[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), "\"'")
	}
	return body, fm
}

func firstHeading(body string) string {
	for _, line := range strings.Split(body, "\n") {
		l := strings.TrimSpace(line)
		if strings.HasPrefix(l, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(l, "#"))
		}
	}
	return ""
}

func slugFromPath(rel, override string) string {
	if override != "" {
		return strings.TrimPrefix(override, "/")
	}
	s := strings.TrimSuffix(rel, filepath.Ext(rel))
	s = strings.ReplaceAll(s, string(filepath.Separator), "/")
	if strings.HasSuffix(s, "/intro") || s == "intro" {
		return ""
	}
	return s
}

func joinURL(base, slug string) string {
	if base == "" {
		return ""
	}
	base = strings.TrimRight(base, "/")
	if slug == "" {
		return base
	}
	return base + "/" + slug
}
