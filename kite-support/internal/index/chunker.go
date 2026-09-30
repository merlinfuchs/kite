package index

import (
	"strings"
)

type Chunk struct {
	Source  string
	Title   string
	URL     string
	Heading string
	Content string
}

func BuildChunks(docs []Document, size, overlap int) []Chunk {
	var out []Chunk
	for _, d := range docs {
		sections := splitByHeadings(d.Body)
		for _, sec := range sections {
			pieces := splitSized(sec.body, size, overlap)
			for _, p := range pieces {
				if strings.TrimSpace(p) == "" {
					continue
				}
				out = append(out, Chunk{
					Source:  d.Source,
					Title:   d.Title,
					URL:     d.URL,
					Heading: sec.heading,
					Content: p,
				})
			}
		}
	}
	return out
}

type section struct {
	heading string
	body    string
}

func splitByHeadings(body string) []section {
	var secs []section
	cur := section{heading: ""}
	var b strings.Builder
	for _, line := range strings.Split(body, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "## ") || strings.HasPrefix(t, "### ") {
			cur.body = b.String()
			if strings.TrimSpace(cur.body) != "" {
				secs = append(secs, cur)
			}
			cur = section{heading: strings.TrimSpace(strings.TrimLeft(t, "# "))}
			b.Reset()
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	cur.body = b.String()
	if strings.TrimSpace(cur.body) != "" {
		secs = append(secs, cur)
	}
	return secs
}

func splitSized(text string, size, overlap int) []string {
	if size <= 0 {
		size = 3000
	}
	if overlap < 0 || overlap >= size {
		overlap = size / 10
	}
	if len(text) <= size {
		return []string{text}
	}

	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	var cur strings.Builder
	for _, p := range paragraphs {
		if cur.Len()+len(p)+2 > size && cur.Len() > 0 {
			chunks = append(chunks, cur.String())
			tail := tailRunes(cur.String(), overlap)
			cur.Reset()
			cur.WriteString(tail)
			if tail != "" {
				cur.WriteString("\n\n")
			}
		}
		cur.WriteString(p)
		cur.WriteString("\n\n")
	}
	if strings.TrimSpace(cur.String()) != "" {
		chunks = append(chunks, cur.String())
	}
	return chunks
}

func tailRunes(s string, n int) string {
	if n <= 0 || len(s) == 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
