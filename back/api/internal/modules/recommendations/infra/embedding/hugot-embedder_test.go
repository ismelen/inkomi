//go:build embedding

// Runs the real model: go test -tags embedding ./internal/modules/recommendations/infra/embedding/
// (downloads multilingual-e5-small, ~120 MB, into EMBEDDING_TEST_DIR or a temp dir).
package embedding_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/infra/embedding"
)

func TestHugotEmbedder_EmbedPassages_SpanishBooks_KeepsSimilarOnesClose(t *testing.T) {
	// Arrange
	dir := os.Getenv("EMBEDDING_TEST_DIR")
	if dir == "" {
		dir = t.TempDir()
	}
	embedder := embedding.NewHugotEmbedder(embedding.NewModelFiles(dir,
		"https://huggingface.co/Xenova/multilingual-e5-small/resolve/main", "onnx/model_quantized.onnx"))
	defer embedder.Release()

	// Act
	vectors, err := embedder.EmbedPassages(context.Background(), []string{
		"Cien años de soledad · Gabriel García Márquez · realismo mágico, Colombia, familia",
		"El amor en los tiempos del cólera · Gabriel García Márquez · amor, Colombia",
		"Dune · Frank Herbert · science fiction, desert planet, politics",
	})

	// Assert
	require.NoError(t, err)
	require.Len(t, vectors, 3)
	assert.Len(t, vectors[0], 384)
	assert.Greater(t, vectors[0].Cosine(vectors[1]), vectors[0].Cosine(vectors[2]))
}
