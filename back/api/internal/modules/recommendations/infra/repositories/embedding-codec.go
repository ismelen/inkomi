package repositories

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

// encodeEmbedding stores a vector as little-endian float32s.
func encodeEmbedding(embedding models.Embedding) []byte {
	data := make([]byte, 4*len(embedding))
	for i, value := range embedding {
		binary.LittleEndian.PutUint32(data[4*i:], math.Float32bits(value))
	}
	return data
}

func decodeEmbedding(data []byte) (models.Embedding, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("blob of %d bytes is not a float32 vector", len(data))
	}
	embedding := make(models.Embedding, len(data)/4)
	for i := range embedding {
		embedding[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[4*i:]))
	}
	return embedding, nil
}
