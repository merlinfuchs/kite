package knowledge

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Doc struct {
	Title string
	URL   string
	Body  string
	// Block types whose settings are described in Body.
	Blocks []string
}

var (
	importLine    = regexp.MustCompile(`(?m)^import \w+ from "[^"]+";?$\n?`)
	embedFlowNode = regexp.MustCompile(`<EmbedFlowNode [^>]*/>\n?`)
	nodeInfo      = regexp.MustCompile(`<NodeInfoExplorer type="([a-z_]+)" />`)
	image         = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)\n?`)
	admonition    = regexp.MustCompile(`(?m)^:::.*$\n?`)
	link          = regexp.MustCompile(`\]\(([^)\s]+)\)`)
	blankLines    = regexp.MustCompile(`\n{3,}`)
)

// LoadDocs reads the Docusaurus docs in root and turns them into plain
// markdown. The block pages only embed a component that shows the block's
// settings, so it's replaced with the same information from the catalog.
func LoadDocs(root, baseURL string, catalog *Catalog) ([]Doc, error) {
	var docs []Doc
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || (filepath.Ext(path) != ".md" && filepath.Ext(path) != ".mdx") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		body, fm := splitFrontmatter(string(raw))

		var blocks []string
		body = nodeInfo.ReplaceAllStringFunc(body, func(m string) string {
			nodeType := nodeInfo.FindStringSubmatch(m)[1]
			block, ok := catalog.Describe(nodeType)
			if !ok {
				slog.Warn("docs page describes an unknown block", "page", rel, "type", nodeType)
				return ""
			}
			blocks = append(blocks, nodeType)
			return block
		})
		body = importLine.ReplaceAllString(body, "")
		body = embedFlowNode.ReplaceAllString(body, "")
		body = image.ReplaceAllString(body, "")
		body = link.ReplaceAllStringFunc(body, func(m string) string {
			return "](" + resolveLink(rel, link.FindStringSubmatch(m)[1], baseURL) + ")"
		})
		body = admonition.ReplaceAllString(body, "")
		body = blankLines.ReplaceAllString(strings.TrimSpace(body), "\n\n")

		title := fm["title"]
		if title == "" {
			title = firstHeading(body)
		}
		if title == "" {
			title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		}

		docs = append(docs, Doc{
			Title:  title,
			URL:    joinURL(baseURL, slugFromPath(rel, fm["slug"])),
			Body:   body,
			Blocks: blocks,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].URL < docs[j].URL })
	return docs, nil
}

// resolveLink turns a link to another page of the docs into the page's URL,
// so the model can link it.
func resolveLink(rel, target, baseURL string) string {
	if strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") || strings.HasPrefix(target, "#") {
		return target
	}
	target, anchor, _ := strings.Cut(target, "#")
	if !strings.HasPrefix(target, "/") {
		target = path.Join(path.Dir(filepath.ToSlash(rel)), target)
	}
	url := joinURL(baseURL, slugFromPath(strings.TrimPrefix(target, "/"), ""))
	if anchor != "" {
		url += "#" + anchor
	}
	return url
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
	for _, line := range strings.Split(s[3:3+end], "\n") {
		k, v, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		fm[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), "\"'")
	}
	return s[3+end+3:], fm
}

func firstHeading(body string) string {
	for _, line := range strings.Split(body, "\n") {
		if l := strings.TrimSpace(line); strings.HasPrefix(l, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(l, "#"))
		}
	}
	return ""
}

func slugFromPath(rel, override string) string {
	if override != "" {
		return strings.Trim(override, "/")
	}
	s := strings.TrimSuffix(filepath.ToSlash(rel), filepath.Ext(rel))
	if s == "index" {
		return ""
	}
	return strings.TrimSuffix(s, "/index")
}

func joinURL(base, slug string) string {
	base = strings.TrimRight(base, "/")
	if slug == "" {
		return base
	}
	return base + "/" + slug
}
