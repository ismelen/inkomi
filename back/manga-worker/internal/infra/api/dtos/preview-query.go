package dtos

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
)

// submittedParam marks a query sent by the form: only then an absent checkbox means false.
const submittedParam = "submitted"

type previewField struct {
	name  string // query param
	label string
	path  []string // path in the MangaConfig JSON
	bool  bool
}

// previewFields are the MangaConfig fields the preview can tune. RotateDoublePage is left out
// because the preview never splits or rotates.
var previewFields = []previewField{
	{name: "targetWidth", label: "Ancho (px)", path: []string{"targetWidth"}},
	{name: "targetHeight", label: "Alto (px)", path: []string{"targetHeight"}},
	{name: "cropTolerance", label: "Recorte de márgenes [0-1], -1 desactivado", path: []string{"cropTolerance"}},
	{name: "enlarge", label: "Ampliar imágenes pequeñas", path: []string{"enlarge"}, bool: true},
	{name: "embed", label: "Rellenar hasta el tamaño", path: []string{"embed"}, bool: true},
	{name: "gamma", label: "Gamma [0.1-3]", path: []string{"gamma"}},
	{name: "brightness", label: "Brillo [-1-1]", path: []string{"brightness"}},
	{name: "contrast", label: "Contraste [0-2]", path: []string{"contrast"}},
	{name: "grayscale", label: "Escala de grises", path: []string{"grayscale"}, bool: true},
	{name: "blackPointThreshold", label: "Punto negro [0-1]", path: []string{"blackPointThreshold"}},
	{name: "whitePointThreshold", label: "Punto blanco [0-1]", path: []string{"whitePointThreshold"}},
	{name: "dithering", label: "Dithering", path: []string{"dithering"}, bool: true},
	{name: "palette8bit", label: "Paleta de 8 bits (PNG)", path: []string{"palette8bit"}, bool: true},
	{name: "quality", label: "Calidad [1-100]", path: []string{"quality"}},
	{name: "sharpenEnabled", label: "Enfoque", path: []string{"sharpen", "enabled"}, bool: true},
	{name: "sharpenRadius", label: "Enfoque: radio [1-5]", path: []string{"sharpen", "radius"}},
	{name: "sharpenX1", label: "Enfoque: x1 [1-3]", path: []string{"sharpen", "x1"}},
	{name: "sharpenY2", label: "Enfoque: y2 [0-50]", path: []string{"sharpen", "y2"}},
	{name: "sharpenY3", label: "Enfoque: y3 [0-50]", path: []string{"sharpen", "y3"}},
	{name: "sharpenM1", label: "Enfoque: m1 [0-]", path: []string{"sharpen", "m1"}},
	{name: "sharpenM2", label: "Enfoque: m2 [1-5]", path: []string{"sharpen", "m2"}},
}

type PreviewFormField struct {
	Name    string
	Label   string
	Value   string
	Bool    bool
	Checked bool
}

// PreviewForm holds the values to repaint the form as the user sent them.
type PreviewForm struct {
	Sample string
	Fields []PreviewFormField
}

type PreviewRequest struct {
	Sample models.PreviewSample
	Config models.MangaConfig
	Form   PreviewForm
}

// ParsePreviewQuery reads the sample and the MangaConfig of a preview query. Empty fields keep
// their default value. The returned request always carries Form, even with an error, so the page
// can be repainted with what the user sent.
func ParsePreviewQuery(values url.Values) (*PreviewRequest, error) {
	req := &PreviewRequest{Form: newPreviewForm(values)}

	sample, err := models.NewPreviewSample(req.Form.Sample)
	if err != nil {
		return req, err
	}
	req.Sample = sample

	raw, err := configJSON(values)
	if err != nil {
		return req, fmt.Errorf("%w: %w", models.ErrInvalidMangaConfig, err)
	}
	req.Config, err = models.NewMangaConfig(raw)
	return req, err
}

// Query encodes the form back as a query, as the form itself would submit it.
func (f *PreviewForm) Query() string {
	values := url.Values{}
	values.Set("sample", f.Sample)
	values.Set(submittedParam, "1")
	for _, field := range f.Fields {
		switch {
		case field.Bool && field.Checked:
			values.Set(field.Name, "1")
		case !field.Bool && field.Value != "":
			values.Set(field.Name, field.Value)
		}
	}
	return values.Encode()
}

func newPreviewForm(values url.Values) PreviewForm {
	defaults := defaultValues()
	submitted := values.Get(submittedParam) != ""

	form := PreviewForm{Sample: values.Get("sample"), Fields: make([]PreviewFormField, len(previewFields))}
	if form.Sample == "" {
		form.Sample = string(models.PreviewSampleBW)
	}
	for i, f := range previewFields {
		field := PreviewFormField{Name: f.name, Label: f.label, Bool: f.bool}
		value := strings.TrimSpace(values.Get(f.name))
		switch {
		case f.bool && submitted:
			field.Checked = value != ""
		case f.bool:
			field.Checked = defaults[f.name] == "true"
		case value != "":
			field.Value = value
		default:
			field.Value = defaults[f.name]
		}
		form.Fields[i] = field
	}
	return form
}

func configJSON(values url.Values) ([]byte, error) {
	submitted := values.Get(submittedParam) != ""
	cfg := map[string]any{}
	for _, f := range previewFields {
		value := strings.TrimSpace(values.Get(f.name))
		switch {
		case f.bool && submitted:
			setPath(cfg, f.path, value != "")
		case f.bool, value == "":
			continue
		default:
			if _, err := strconv.ParseFloat(value, 64); err != nil {
				return nil, fmt.Errorf("%s %q is not a number", f.name, value)
			}
			setPath(cfg, f.path, json.Number(value))
		}
	}
	return json.Marshal(cfg)
}

func setPath(m map[string]any, path []string, value any) {
	for _, key := range path[:len(path)-1] {
		child, ok := m[key].(map[string]any)
		if !ok {
			child = map[string]any{}
			m[key] = child
		}
		m = child
	}
	m[path[len(path)-1]] = value
}

// defaultValues formats DefaultMangaConfig by query param name.
func defaultValues() map[string]string {
	raw, err := json.Marshal(models.DefaultMangaConfig())
	if err != nil {
		return nil
	}
	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil
	}

	defaults := make(map[string]string, len(previewFields))
	for _, f := range previewFields {
		var value any = cfg
		for _, key := range f.path {
			m, ok := value.(map[string]any)
			if !ok {
				value = nil
				break
			}
			value = m[key]
		}
		switch v := value.(type) {
		case float64:
			defaults[f.name] = strconv.FormatFloat(v, 'f', -1, 64)
		case bool:
			defaults[f.name] = strconv.FormatBool(v)
		}
	}
	return defaults
}
