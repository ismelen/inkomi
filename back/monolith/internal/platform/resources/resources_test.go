package resources_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ismelen/inkomi/back/monolith/internal/platform/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mb = 1 << 20

func files(t *testing.T, meminfoKB, cgroupMax, cgroupCurrent string) *resources.Memory {
	t.Helper()
	dir := t.TempDir()
	meminfo := filepath.Join(dir, "meminfo")
	if meminfoKB != "" {
		require.NoError(t, os.WriteFile(meminfo, []byte("MemTotal: 4000000 kB\nMemAvailable: "+meminfoKB+" kB\n"), 0o600))
	}
	cgroup := filepath.Join(dir, "cgroup")
	require.NoError(t, os.Mkdir(cgroup, 0o700))
	if cgroupMax != "" {
		require.NoError(t, os.WriteFile(filepath.Join(cgroup, "memory.max"), []byte(cgroupMax+"\n"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(cgroup, "memory.current"), []byte(cgroupCurrent+"\n"), 0o600))
	}
	return resources.NewMemoryAt(meminfo, cgroup)
}

func TestMemory_Available_Sources(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		meminfoKB string
		max       string
		current   string
		want      uint64
		wantOk    bool
	}{
		{"Machine only", "2048000", "", "", 2048000 * 1024, true},
		{"No cgroup limit", "2048000", "max", "0", 2048000 * 1024, true},
		{"Cgroup limit is lower", "2048000", "1073741824", "536870912", 512 * mb, true},
		{"Machine is lower", "102400", "1073741824", "0", 100 * mb, true},
		{"Nothing readable", "", "", "", 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			memory := files(t, tc.meminfoKB, tc.max, tc.current)

			// Act
			got, ok := memory.Available()

			// Assert
			assert.Equal(t, tc.wantOk, ok)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestConcurrency_Limits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		cpus      int
		available uint64
		ok        bool
		want      int
	}{
		{"Enough memory", 4, 4096 * mb, true, 4},
		{"Memory bound", 4, 1024 * mb, true, 2},
		{"Never zero", 4, 100 * mb, true, 1},
		{"Unknown memory", 2, 0, false, 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			got := resources.Concurrency(tc.cpus, tc.available, tc.ok, 512*mb)

			// Assert
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestMemoryGate_Ready(t *testing.T) {
	t.Parallel()

	// Arrange
	low := resources.NewMemoryGate(files(t, "204800", "", ""), 400*mb)
	high := resources.NewMemoryGate(files(t, "2048000", "", ""), 400*mb)
	unknown := resources.NewMemoryGate(files(t, "", "", ""), 400*mb)

	// Act & Assert
	assert.False(t, low.Ready())
	assert.True(t, high.Ready())
	assert.True(t, unknown.Ready())
}
