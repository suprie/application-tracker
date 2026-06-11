package rag

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// BM25 parameters — standard defaults from the research literature.
const (
	bm25k1 = 1.5 // term frequency saturation
	bm25b  = 0.75 // length normalization
)

// bm25Index holds precomputed per-chunk term frequencies and corpus-wide
// document frequencies so the BM25 score can be computed without an
// embedding model.
type bm25Index struct {
	chunks    []Chunk
	chunkTF   []map[string]int // chunkTF[i] = term → frequency in chunk i
	docFreq   map[string]int   // docFreq[t] = number of chunks containing t
	avgLen    float64          // average number of tokens per chunk
	chunkLens []int            // token count per chunk
}

// newBM25Index tokenizes every chunk and builds the inverted index.
func newBM25Index(chunks []Chunk) *bm25Index {
	idx := &bm25Index{
		chunks:    chunks,
		chunkTF:   make([]map[string]int, len(chunks)),
		docFreq:   make(map[string]int),
		chunkLens: make([]int, len(chunks)),
	}

	var totalLen int
	seenInDoc := make(map[string]bool)

	for i, c := range chunks {
		tokens := tokenize(c.Content)
		idx.chunkLens[i] = len(tokens)
		totalLen += len(tokens)

		tf := make(map[string]int, len(tokens))
		for _, t := range tokens {
			tf[t]++
		}
		idx.chunkTF[i] = tf

		// Document frequency: count each term once per chunk.
		for t := range tf {
			if !seenInDoc[t] {
				idx.docFreq[t]++
				seenInDoc[t] = true
			}
		}
		for t := range seenInDoc {
			delete(seenInDoc, t)
		}
	}

	if len(chunks) > 0 {
		idx.avgLen = float64(totalLen) / float64(len(chunks))
	}
	return idx
}

// topK returns the k chunks with the highest BM25 scores for query.
func (idx *bm25Index) topK(query string, k int) []Chunk {
	if len(idx.chunks) == 0 {
		return nil
	}
	if k <= 0 {
		k = 5
	}

	queryTokens := tokenize(query)
	if len(queryTokens) == 0 {
		limit := k
		if limit > len(idx.chunks) {
			limit = len(idx.chunks)
		}
		return idx.chunks[:limit]
	}

	// Compute query-term IDF once.
	N := float64(len(idx.chunks))
	queryIDF := make(map[string]float64, len(queryTokens))
	for _, t := range queryTokens {
		if _, ok := queryIDF[t]; ok {
			continue
		}
		n := float64(idx.docFreq[t])
		queryIDF[t] = math.Log((N-n+0.5)/(n+0.5) + 1.0)
	}

	type scored struct {
		chunk Chunk
		score float64
	}
	scores := make([]scored, len(idx.chunks))

	for i := range idx.chunks {
		var sum float64
		docLen := float64(idx.chunkLens[i])
		tf := idx.chunkTF[i]

		for _, t := range queryTokens {
			f := float64(tf[t])
			if f == 0 {
				continue
			}
			idf := queryIDF[t]
			num := f * (bm25k1 + 1)
			den := f + bm25k1*(1-bm25b+bm25b*docLen/idx.avgLen)
			sum += idf * num / den
		}
		scores[i] = scored{chunk: idx.chunks[i], score: sum}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	limit := k
	if limit > len(scores) {
		limit = len(scores)
	}
	result := make([]Chunk, limit)
	for i := 0; i < limit; i++ {
		result[i] = scores[i].chunk
	}
	return result
}

// tokenize splits text into lowercase tokens, stripping punctuation and
// dropping tokens shorter than 2 characters.
func tokenize(text string) []string {
	raw := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	tokens := make([]string, 0, len(raw))
	for _, t := range raw {
		if len(t) >= 2 {
			tokens = append(tokens, t)
		}
	}
	return tokens
}

// Retrieve runs BM25 keyword retrieval over the chunks in the store.
// It requires no embedding model — the inverted index is built on the fly
// from the chunks already in memory (~30 chunks, microseconds).
func Retrieve(query string, store *ChunkStore, k int) ([]Chunk, error) {
	if store.Len() == 0 {
		return nil, nil
	}
	idx := newBM25Index(store.Chunks())
	return idx.topK(query, k), nil
}
