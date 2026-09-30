package index

import (
	"encoding/gob"
	"fmt"
	"io"
	"math"
	"sort"
)

type storedChunk struct {
	Source    string
	Title     string
	URL       string
	Heading   string
	Content   string
	Embedding []float32
	Norm      float32
}

type Store struct {
	chunks []storedChunk
}

type Hit struct {
	Chunk    Chunk
	Distance float32 // 1 - cosine similarity; lower = better
}

func NewStore() *Store { return &Store{} }

func (s *Store) Insert(chunks []Chunk, vectors [][]float32) error {
	if len(chunks) != len(vectors) {
		return fmt.Errorf("chunks/vectors length mismatch: %d vs %d", len(chunks), len(vectors))
	}
	for i, c := range chunks {
		v := vectors[i]
		s.chunks = append(s.chunks, storedChunk{
			Source:    c.Source,
			Title:     c.Title,
			URL:       c.URL,
			Heading:   c.Heading,
			Content:   c.Content,
			Embedding: v,
			Norm:      vectorNorm(v),
		})
	}
	return nil
}

func (s *Store) Encode(w io.Writer) error {
	return gob.NewEncoder(w).Encode(s.chunks)
}

func (s *Store) Decode(r io.Reader) error {
	return gob.NewDecoder(r).Decode(&s.chunks)
}

func (s *Store) Count() int { return len(s.chunks) }

func (s *Store) Search(query []float32, k int) ([]Hit, error) {
	qnorm := vectorNorm(query)
	if qnorm == 0 || len(s.chunks) == 0 {
		return nil, nil
	}

	type scored struct {
		idx int
		sim float32
	}
	scores := make([]scored, len(s.chunks))
	for i, c := range s.chunks {
		if c.Norm == 0 || len(c.Embedding) != len(query) {
			scores[i] = scored{i, -1}
			continue
		}
		scores[i] = scored{i, dot(query, c.Embedding) / (qnorm * c.Norm)}
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].sim > scores[j].sim })
	if k > len(scores) {
		k = len(scores)
	}

	hits := make([]Hit, k)
	for i := 0; i < k; i++ {
		c := s.chunks[scores[i].idx]
		hits[i] = Hit{
			Chunk: Chunk{
				Source:  c.Source,
				Title:   c.Title,
				URL:     c.URL,
				Heading: c.Heading,
				Content: c.Content,
			},
			Distance: 1 - scores[i].sim,
		}
	}
	return hits, nil
}

func vectorNorm(v []float32) float32 {
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	return float32(math.Sqrt(sum))
}

func dot(a, b []float32) float32 {
	var sum float64
	for i := range a {
		sum += float64(a[i]) * float64(b[i])
	}
	return float32(sum)
}
