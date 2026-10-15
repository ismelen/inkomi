package epub

import "text/template"

const (
	containerXML = `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>
`

	opfTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="pub-id"{{with .Prefix}} prefix="{{xml .}}"{{end}}>
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
    <dc:identifier id="pub-id">urn:uuid:{{.UUID}}</dc:identifier>
    <dc:title>{{xml .Title}}</dc:title>
    {{- with .Author}}
    <dc:creator>{{xml .}}</dc:creator>
    {{- end}}
    <dc:language>{{xml .Language}}</dc:language>
    <meta property="dcterms:modified">{{.Modified}}</meta>
    {{- with .CoverId}}
    <meta name="cover" content="{{.}}"/>
    {{- end}}
    <meta name="primary-writing-mode" content="{{.WritingMode}}"/>
    {{- range .Metas}}
    {{- if .Property}}
    <meta property="{{xml .Property}}">{{xml .Value}}</meta>
    {{- else}}
    <meta name="{{xml .Name}}" content="{{xml .Content}}"/>
    {{- end}}
    {{- end}}
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
    {{- range .Manifest}}
    <item id="{{.Id}}" href="{{xml .Href}}" media-type="{{xml .MediaType}}"{{with .Properties}} properties="{{xml .}}"{{end}}/>
    {{- end}}
  </manifest>
  <spine toc="ncx" page-progression-direction="{{.Direction}}">
    {{- range .Spine}}
    <itemref idref="{{.IdRef}}"{{with .Linear}} linear="{{xml .}}"{{end}}{{with .Properties}} properties="{{xml .}}"{{end}}/>
    {{- end}}
  </spine>
</package>
`

	navTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head><title>{{xml .Title}}</title></head>
<body>
  <nav epub:type="toc" id="toc">
    <h1>{{xml .Title}}</h1>
    <ol>
    {{- range .Toc}}
      <li><a href="{{xml .Href}}">{{xml .Title}}</a></li>
    {{- end}}
    </ol>
  </nav>
</body>
</html>
`

	ncxTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1">
  <head>
    <meta name="dtb:uid" content="urn:uuid:{{.UUID}}"/>
  </head>
  <docTitle><text>{{xml .Title}}</text></docTitle>
  <navMap>
  {{- range $i, $e := .Toc}}
    <navPoint id="navpoint-{{inc $i}}" playOrder="{{inc $i}}">
      <navLabel><text>{{xml $e.Title}}</text></navLabel>
      <content src="{{xml $e.Href}}"/>
    </navPoint>
  {{- end}}
  </navMap>
</ncx>
`
)

var templates = template.Must(template.New("epub").Funcs(template.FuncMap{
	"xml": xmlEscape,
	"inc": func(i int) int { return i + 1 },
}).Parse(`{{define "opf"}}` + opfTemplate + `{{end}}{{define "nav"}}` + navTemplate + `{{end}}{{define "ncx"}}` + ncxTemplate + `{{end}}`))
