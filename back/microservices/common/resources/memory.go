package resources

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Memory reads the free memory of the machine and, inside a container, of its cgroup limit.
type Memory struct {
	meminfo string // /proc/meminfo
	cgroup  string // cgroup v2 mount, /sys/fs/cgroup
}

func NewMemory() *Memory {
	return &Memory{meminfo: "/proc/meminfo", cgroup: "/sys/fs/cgroup"}
}

// NewMemoryAt reads the given files instead of the system ones.
func NewMemoryAt(meminfo, cgroup string) *Memory {
	return &Memory{meminfo: meminfo, cgroup: cgroup}
}

// Available returns the bytes that can still be allocated: MemAvailable of the machine, capped
// by what is left under the cgroup limit when there is one. ok is false when neither can be
// read (e.g. outside Linux).
func (m *Memory) Available() (bytes uint64, ok bool) {
	machine, machineOk := m.memAvailable()
	limit, limitOk := m.cgroupFree()
	switch {
	case machineOk && limitOk:
		return min(machine, limit), true
	case machineOk:
		return machine, true
	case limitOk:
		return limit, true
	default:
		return 0, false
	}
}

func (m *Memory) memAvailable() (uint64, bool) {
	f, err := os.Open(m.meminfo)
	if err != nil {
		return 0, false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[0] == "MemAvailable:" {
			kb, err := strconv.ParseUint(fields[1], 10, 64)
			return kb * 1024, err == nil
		}
	}
	return 0, false
}

func (m *Memory) cgroupFree() (uint64, bool) {
	limit, ok := readUint(filepath.Join(m.cgroup, "memory.max"))
	if !ok {
		return 0, false
	}
	used, ok := readUint(filepath.Join(m.cgroup, "memory.current"))
	if !ok || used >= limit {
		return 0, ok
	}
	return limit - used, true
}

// readUint reads a number file; "max" (no limit) counts as unreadable.
func readUint(path string) (uint64, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	n, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	return n, err == nil
}
