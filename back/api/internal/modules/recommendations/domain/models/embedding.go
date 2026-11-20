package models

import "math"

// Embedding is a vector of the meaning of a work; the embedder returns them L2-normalized.
type Embedding []float32

// Cosine is the cosine similarity of e and other, 0 when either is empty or of other size.
func (e Embedding) Cosine(other Embedding) float64 {
	if len(e) == 0 || len(e) != len(other) {
		return 0
	}
	var dot, a, b float64
	for i := range e {
		dot += float64(e[i]) * float64(other[i])
		a += float64(e[i]) * float64(e[i])
		b += float64(other[i]) * float64(other[i])
	}
	if a == 0 || b == 0 {
		return 0
	}
	return dot / math.Sqrt(a*b)
}

// WeightedMean is the normalized weighted mean of vectors, or nil when there are none.
func WeightedMean(vectors []Embedding, weights []float64) Embedding {
	if len(vectors) == 0 || len(vectors) != len(weights) {
		return nil
	}
	sum := make([]float64, len(vectors[0]))
	for i, vector := range vectors {
		if len(vector) != len(sum) {
			continue
		}
		for j, value := range vector {
			sum[j] += float64(value) * weights[i]
		}
	}
	var norm float64
	for _, value := range sum {
		norm += value * value
	}
	if norm == 0 {
		return nil
	}
	norm = math.Sqrt(norm)
	mean := make(Embedding, len(sum))
	for i, value := range sum {
		mean[i] = float32(value / norm)
	}
	return mean
}
