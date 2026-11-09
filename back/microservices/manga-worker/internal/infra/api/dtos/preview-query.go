package dtos

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ismelen/inkomi/back/microservices/manga-worker/internal/domain/models"
)

// submittedParam marks a query sent by the form: only then an absent checkbox means false.
const submittedParam = "submitted"

type previewField struct {
	name  string // query param
	label string
	group string   // fieldset of the form
	path  []string // path in the MangaConfig JSON
	bool  bool
}

const (
	groupSize    = "Tamaño"
	groupTone    = "Tono"
	groupOutput  = "Salida"
	groupSharpen = "Enfoque"
)

// previewFields are the MangaConfig fields the preview can tune, in form order. RotateDoublePage
// is left out because the preview never splits or rotates.
var previewFields = []previewField{
	{name: "targetWidth", label: "Ancho (px)", group: groupSize, path: []string{"targetWidth"}},
	{name: "targetHeight", label: "Alto (px)", group: groupSize, path: []string{"targetHeight"}},
	{name: "cropTolerance", label: "Recorte márgenes [0-1], -1 no", group: groupSize, path: []string{"cropTolerance"}},
	{name: "enlarge", label: "Ampliar pequeñas", group: groupSize, path: []string{"enlarge"}, bool: true},
	{name: "embed", label: "Rellenar hasta el tamaño", group: groupSize, path: []string{"embed"}, bool: true},
	{name: "gamma", label: "Gamma [0.1-3]", group: groupTone, path: []string{"gamma"}},
	{name: "brightness", label: "Brillo [-1-1]", group: groupTone, path: []string{"brightness"}},
	{name: "contrast", label: "Contraste [0-2]", group: groupTone, path: []string{"contrast"}},
	{name: "blackPointThreshold", label: "Punto negro [0-1]", group: groupTone, path: []string{"blackPointThreshold"}},
	{name: "whitePointThreshold", label: "Punto blanco [0-1]", group: groupTone, path: []string{"whitePointThreshold"}},
	{name: "grayscale", label: "Escala de grises", group: groupTone, path: []string{"grayscale"}, bool: true},
	{name: "quality", label: "Calidad [1-100]", group: groupOutput, path: []string{"quality"}},
	{name: "dithering", label: "Dithering", group: groupOutput, path: []string{"dithering"}, bool: true},
	{name: "palette8bit", label: "Paleta 8 bits (PNG)", group: groupOutput, path: []string{"palette8bit"}, bool: true},
	{name: "sharpenEnabled", label: "Activar enfoque", group: groupSharpen, path: []string{"sharpen", "enabled"}, bool: true},
	{name: "sharpenRadius", label: "Radio [1-5]", group: groupSharpen, path: []string{"sharpen", "radius"}},
	{name: "sharpenX1", label: "x1 [1-3]", group: groupSharpen, path: []string{"sharpen", "x1"}},
	{name: "sharpenY2", label: "y2 [0-50]", group: groupSharpen, path: []string{"sharpen", "y2"}},
	{name: "sharpenY3", label: "y3 [0-50]", group: groupSharpen, path: []string{"sharpen", "y3"}},
	{name: "sharpenM1", label: "m1 [0-]", group: groupSharpen, path: []string{"sharpen", "m1"}},
	{name: "sharpenM2", label: "m2 [1-5]", group: groupSharpen, path: []string{"sharpen", "m2"}},
}

type PreviewFormField struct {
	Name    string
	Label   string
	Group   string
	Value   string
	Bool    bool
	Checked bool
}

type PreviewFieldGroup struct {
	Title  string
	Fields []PreviewFormField
}

// PreviewForm holds the values to repaint the form as the user sent them.
type PreviewForm struct {
	Sample string
	Fields []PreviewFormField
}

// Groups splits the fields into the fieldsets of the form, keeping their order.
func (f *PreviewForm) Groups() []PreviewFieldGroup {
	var groups []PreviewFieldGroup
	for _, field := range f.Fields {
		if len(groups) == 0 || groups[len(groups)-1].Title != field.Group {
			groups = append(groups, PreviewFieldGroup{Title: field.Group})
		}
		last := &groups[len(groups)-1]
		last.Fields = append(last.Fields, field)
	}
	return groups
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
		field := PreviewFormField{Name: f.name, Label: f.label, Group: f.group, Bool: f.bool}
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
