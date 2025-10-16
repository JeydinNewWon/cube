package node

type Node struct {
	Name            string
	Ip              string
	Cores           uint
	Memory          uint64
	MemoryAllocated uint64
	Disk            int
	DiskAllocated   int
	Role            string
	TaskCount       int
}
