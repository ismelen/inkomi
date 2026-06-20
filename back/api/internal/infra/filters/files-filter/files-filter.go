package filesFilter

type FilesFilter interface {
	setNext(filter FilesFilter)
	Filter(files []string) (bool, string)
}

type FilesFilterBase struct {
	next FilesFilter
}

func (f *FilesFilterBase) setNext(filter FilesFilter) {
	f.next = filter
}

func (f *FilesFilterBase) Next(files []string) (bool, string) {
	if f.next != nil {
		return f.next.Filter(files)
	}
	return false, ""
}

func Use(filters ...FilesFilter) FilesFilter {
	fst := filters[0]
	prevFilter := filters[0]
	for _, filter := range filters[1:] {
		prevFilter.setNext(filter)
		prevFilter = filter
	}
	return fst
}
