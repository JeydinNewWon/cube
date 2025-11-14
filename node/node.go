package node

type Node struct {
	Name            string
	Api             string
	Ip              string
	Cores           uint
	Memory          uint64
	MemoryAllocated uint64
	Disk            int
	DiskAllocated   int
	Role            string
	TaskCount       int
}

func NewNode(name string, api string, role string) *Node {
	return &Node{
		Name: name,
		Api:  api,
		Role: role,
	}
}
