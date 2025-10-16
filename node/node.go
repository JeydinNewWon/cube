package node

type Node struct {
	name            string
	ip              string
	cores           uint
	memory          uint64
	memoryAllocated uint64
	disk            int
	diskAllocated   int
	role            string
	taskCount       int
}
