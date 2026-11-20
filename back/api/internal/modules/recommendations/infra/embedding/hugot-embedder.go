package embedding

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"

	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/backends"
	"github.com/knights-analytics/hugot/pipelines"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/ports"
)

// passagePrefix is how the e5 models expect a document to be introduced.
const passagePrefix = "passage: "

// HugotEmbedder runs a sentence-embedding ONNX model (multilingual-e5-small) inside the process
// with hugot's pure Go backend: no cgo and no extra library in the image. The model is loaded
// on the first call and freed by Release, since it takes a few hundred MB.
type HugotEmbedder struct {
	model *ModelFiles

	mu       sync.Mutex
	session  *hugot.Session
	pipeline *pipelines.FeatureExtractionPipeline
}

var _ ports.IEmbedder = (*HugotEmbedder)(nil)

func NewHugotEmbedder(model *ModelFiles) *HugotEmbedder {
	return &HugotEmbedder{model: model}
}

func (h *HugotEmbedder) EmbedPassages(ctx context.Context, texts []string) ([]models.Embedding, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.load(ctx); err != nil {
		return nil, err
	}

	inputs := make([]string, len(texts))
	for i, text := range texts {
		inputs[i] = passagePrefix + text
	}
	output, err := h.pipeline.RunPipeline(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("embed %d passages: %w", len(texts), err)
	}
	if len(output.Embeddings) != len(texts) {
		return nil, fmt.Errorf("embed %d passages: got %d vectors", len(texts), len(output.Embeddings))
	}
	result := make([]models.Embedding, len(output.Embeddings))
	for i, vector := range output.Embeddings {
		result[i] = vector
	}
	return result, nil
}

func (h *HugotEmbedder) Release() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.session == nil {
		return
	}
	if err := h.session.Destroy(); err != nil {
		log.Printf("recommendations: release embedding model: %v", err)
	}
	h.session, h.pipeline = nil, nil
	// the model is a few hundred MB: give them back now, the manga jobs need them
	debug.FreeOSMemory()
}

func (h *HugotEmbedder) load(ctx context.Context) error {
	if h.pipeline != nil {
		return nil
	}
	dir, err := h.model.Ensure(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", models.ErrNoModel, err)
	}
	session, err := hugot.NewGoSession(ctx)
	if err != nil {
		return fmt.Errorf("start embedding session: %w", err)
	}
	pipeline, err := hugot.NewPipeline(session, hugot.FeatureExtractionConfig{
		ModelPath:    dir,
		Name:         "embeddings",
		OnnxFilename: h.model.OnnxFilename(),
		Options:      []backends.PipelineOption[*pipelines.FeatureExtractionPipeline]{pipelines.WithNormalization()},
	})
	if err != nil {
		if destroyErr := session.Destroy(); destroyErr != nil {
			log.Printf("recommendations: destroy embedding session: %v", destroyErr)
		}
		return fmt.Errorf("load embedding model: %w", err)
	}
	h.session, h.pipeline = session, pipeline
	return nil
}
