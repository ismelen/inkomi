package models

// SourceDetail is a source together with the related rows needed to dispatch it to a worker.
type SourceDetail struct {
	Source
	Config *Config
	Parent *SourceParent
}

// SourceParent holds the fields of the parent folder that affect its children.
type SourceParent struct {
	ID    string
	Title string
	IsRTL bool
	Join  bool
}
