package handlers

import (
	"context"
	"embed"
	"errors"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/api/dtos"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/getsample"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/previewpage"
)

//go:embed templates/preview.html
var templates embed.FS

var previewTemplate = template.Must(template.ParseFS(templates, "templates/preview.html"))

// cacheControl lets browsers keep previews: the same URL always produces the same image.
const cacheControl = "public, max-age=86400"

var sampleLabels = []struct {
	sample models.PreviewSample
	label  string
}{
	{models.PreviewSampleBW, "Blanco y negro"},
	{models.PreviewSampleColor, "Color"},
}

type sampleOption struct {
	Value    string
	Label    string
	Selected bool
}

type previewPageData struct {
	Form         dtos.PreviewForm
	Samples      []sampleOption
	Error        string
	OriginalURL  string
	ProcessedURL string
}

// PreviewHandler serves the public page to try a MangaConfig on a sample page. It writes HTML
// and images instead of JSON, so it doesn't go through a JSON wrapper.
type PreviewHandler struct {
	previewUC   *previewpage.PreviewPageUC
	getSampleUC *getsample.GetSampleUC
}

func NewPreviewHandler(previewUC *previewpage.PreviewPageUC, getSampleUC *getsample.GetSampleUC) *PreviewHandler {
	return &PreviewHandler{previewUC: previewUC, getSampleUC: getSampleUC}
}

// HandlePage serves GET /preview: the form prefilled from the query, the original sample and,
// when the config is valid, the sample processed with it.
func (p *PreviewHandler) HandlePage(w http.ResponseWriter, r *http.Request) {
	req, err := dtos.ParsePreviewQuery(r.URL.Query())
	if err == nil {
		err = req.Config.FitsPreview()
	}

	sample := req.Form.Sample
	if _, sampleErr := models.NewPreviewSample(sample); sampleErr != nil {
		sample = string(models.PreviewSampleBW)
	}

	data := previewPageData{
		Form:        req.Form,
		OriginalURL: "/preview/original?" + url.Values{"sample": {sample}}.Encode(),
	}
	for _, s := range sampleLabels {
		data.Samples = append(data.Samples, sampleOption{Value: string(s.sample), Label: s.label, Selected: string(s.sample) == sample})
	}
	if err != nil {
		data.Error = err.Error()
	} else {
		data.ProcessedURL = "/preview/processed?" + req.Form.Query()
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := previewTemplate.Execute(w, data); err != nil {
		log.Printf("render preview page: %v", err)
	}
}

// HandleOriginal serves GET /preview/original?sample=<sample>.
func (p *PreviewHandler) HandleOriginal(w http.ResponseWriter, r *http.Request) {
	sample, err := models.NewPreviewSample(r.URL.Query().Get("sample"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := p.getSampleUC.Execute(r.Context(), sample)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeImage(w, http.DetectContentType(data), data)
}

// HandleProcessed serves GET /preview/processed?sample=<sample>&<config>.
func (p *PreviewHandler) HandleProcessed(w http.ResponseWriter, r *http.Request) {
	req, err := dtos.ParsePreviewQuery(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	page, err := p.previewUC.Execute(r.Context(), req.Sample, req.Config)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeImage(w, page.Format.MediaType(), page.Data)
}

func writeImage(w http.ResponseWriter, contentType string, data []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Cache-Control", cacheControl)
	if _, err := w.Write(data); err != nil {
		log.Printf("write preview image: %v", err)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, models.ErrUnknownSample),
		errors.Is(err, models.ErrInvalidMangaConfig),
		errors.Is(err, models.ErrPreviewTooLarge):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, context.DeadlineExceeded):
		http.Error(w, "preview took too long", http.StatusGatewayTimeout)
	case errors.Is(err, context.Canceled):
		// the client is gone, nobody reads the response
	default:
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
