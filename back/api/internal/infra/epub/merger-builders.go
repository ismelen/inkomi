package epub

import (
	"fmt"
	fileBuilder "ismelen/inkomi/internal/infra/epub/file-builder"
	"time"

	"github.com/google/uuid"
)

func buildOPF(title, author string, manifest []mergedItem, spineIDs []string, coverItemID, tocID, pageProgression, writingMode string) string {
	b := fileBuilder.New()
	b.AddFromTemplate(MergerOPFStart, uuid.New().String(), xmlEscape(title), xmlEscape(author), time.Now().UTC().Format("2006-01-02T15:04:05Z"), coverItemID, writingMode)

	for _, m := range manifest {
		props := ""
		if m.Properties != "" {
			props = fmt.Sprintf(` properties="%s"`, m.Properties)
		}
		b.AddFromTemplate(MergerOPFItem, m.ID, m.Href, m.MediaType, props)
	}

	b.AddFromTemplate(MergerOPFSpineStart, tocID, pageProgression)
	for _, id := range spineIDs {
		b.AddFromTemplate(MergerOPFItemRef, id)
	}
	b.Add(MergerOPFEnd)

	return b.Build()
}

func buildNav(title string, entries []navEntry) string {
	b := fileBuilder.New()
	b.AddFromTemplate(MergerNavStart, xmlEscape(title), xmlEscape(title))
	for _, e := range entries {
		if e.Href == "" {
			continue
		}
		b.AddFromTemplate(MergerNavEntry, e.Href, xmlEscape(e.Title))
	}
	b.Add(MergerNavEnd)
	return b.Build()
}

func buildNCX(title string, entries []navEntry) string {
	b := fileBuilder.New()
	b.AddFromTemplate(MergerNCXStart, uuid.New().String(), xmlEscape(title))
	order := 0
	for _, e := range entries {
		if e.Href == "" {
			continue
		}
		order++
		b.AddFromTemplate(MergerNCXEntry, order, order, xmlEscape(e.Title), e.Href)
	}
	b.Add(MergerNCXEnd)
	return b.Build()
}
