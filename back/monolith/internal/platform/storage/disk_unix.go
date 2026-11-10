//go:build unix

package storage

import "syscall"

// FreeBytes returns the bytes an unprivileged process can still write on the filesystem of path.
func FreeBytes(path string) (uint64, bool) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, false
	}
	return stat.Bavail * uint64(stat.Bsize), true
}
