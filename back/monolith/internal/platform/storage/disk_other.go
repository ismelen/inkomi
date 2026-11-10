//go:build !unix

package storage

// FreeBytes can't read the free space outside Unix; callers then don't limit by disk.
func FreeBytes(path string) (uint64, bool) {
	return 0, false
}
