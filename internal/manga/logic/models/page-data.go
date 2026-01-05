package manga

type PageData struct {
	BgColor     string
	Count int8
	Src      string
	Payloads [3]*PagePayload
}

func NewPageData(path string) *PageData {
	return &PageData{Src: path, Count: 1}
}
