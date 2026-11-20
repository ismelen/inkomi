package resources

// Concurrency returns how many jobs that need perJob bytes each can run at once: as many as
// cpus, fewer when the free memory doesn't fit them, and never less than one. Without memory
// information it trusts cpus.
func Concurrency(cpus int, available uint64, availableOk bool, perJob uint64) int {
	n := max(1, cpus)
	if availableOk && perJob > 0 {
		n = min(n, int(available/perJob))
	}
	return max(1, n)
}

// MemoryGate tells whether there is enough free memory to start another job.
type MemoryGate struct {
	memory  *Memory
	reserve uint64
}

func NewMemoryGate(memory *Memory, reserve uint64) *MemoryGate {
	return &MemoryGate{memory: memory, reserve: reserve}
}

// Ready reports whether more than the reserve is free; it is always true when the free memory
// can't be read.
func (g *MemoryGate) Ready() bool {
	available, ok := g.memory.Available()
	return !ok || available > g.reserve
}
