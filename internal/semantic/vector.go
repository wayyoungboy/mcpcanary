package semantic

import (
	"hash/fnv"
	"math"
	"regexp"
	"strings"
)

var tokenPattern = regexp.MustCompile(`[a-z0-9_:/.-]+`)

func Embed(text string, dimension int) []float64 {
	if dimension <= 0 {
		dimension = 128
	}
	vector := make([]float64, dimension)
	tokens := tokenPattern.FindAllString(strings.ToLower(text), -1)
	for i, token := range tokens {
		addToken(vector, token, 1)
		if i+1 < len(tokens) {
			addToken(vector, token+" "+tokens[i+1], 0.6)
		}
	}
	normalize(vector)
	return vector
}

func Cosine(a, b []float64) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, magA, magB float64
	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}
	if magA == 0 || magB == 0 {
		return 0
	}
	return dot / (math.Sqrt(magA) * math.Sqrt(magB))
}

func addToken(vector []float64, token string, weight float64) {
	h := fnv.New64a()
	_, _ = h.Write([]byte(token))
	sum := h.Sum64()
	idx := int(sum % uint64(len(vector)))
	sign := 1.0
	if sum&1 == 1 {
		sign = -1
	}
	vector[idx] += sign * weight
}

func normalize(vector []float64) {
	var mag float64
	for _, value := range vector {
		mag += value * value
	}
	if mag == 0 {
		return
	}
	scale := math.Sqrt(mag)
	for i := range vector {
		vector[i] /= scale
	}
}
