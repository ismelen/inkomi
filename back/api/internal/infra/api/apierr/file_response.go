package apierr

// FileResponse signals the request wrapper to serve a file from disk.
// This type lives in infra/api because it is an HTTP response abstraction.
type FileResponse struct {
	Path, Name string
	Remove     bool
}
