package scheduler

import (
	"cube/node"
	"cube/task"
)

type Scheduler interface {
	SelectCandidateNodes(task task.Task, nodes []*node.Node) []*node.Node
	Score(t task.Task, nodes []*node.Node) map[string]float64
	Pick(scores map[string]float64, nodes []*node.Node) *node.Node
}

type RoundRobin struct {
	Name       string
	LastWorker int
}

func (r *RoundRobin) SelectCandidateNodes(task task.Task, nodes []*node.Node) []*node.Node {
	return nodes
}

func (r *RoundRobin) Score(t task.Task, nodes []*node.Node) map[string]float64 {
	workerMap := make(map[string]float64)
	var newWorker int
	if r.LastWorker+1 < len(nodes) {
		newWorker = r.LastWorker + 1
		r.LastWorker++
	} else {
		newWorker = 0
		r.LastWorker = 0
	}

	for idx, n := range nodes {
		if idx == newWorker {
			workerMap[n.Name] = 0.1
		} else {
			workerMap[n.Name] = 1.0
		}
	}

	return workerMap
}

func (r *RoundRobin) Pick(scores map[string]float64, nodes []*node.Node) *node.Node {
	var bestNode *node.Node
	var lowestScore float64

	for idx, node := range nodes {
		if idx == 0 {
			bestNode = node
			lowestScore = scores[node.Name]
		}

		if scores[node.Name] < lowestScore {
			bestNode = nodes[idx]
			lowestScore = scores[node.Name]
		}
	}

	return bestNode
}
