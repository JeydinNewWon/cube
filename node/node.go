package node

import (
	"cube/stats"
	"cube/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Node struct {
	Name            string
	Api             string
	Ip              string
	Cores           uint
	Memory          uint64
	MemoryAllocated uint64
	Disk            int64
	DiskAllocated   int64
	Stats           stats.Stats
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

func (n *Node) GetStats() (*stats.Stats, error) {
	var resp *http.Response
	var err error

	url := fmt.Sprintf("%s/stats", n.Api)
	resp, err = utils.HttpWithRetry(http.Get, url)
	if err != nil {
		msg := fmt.Sprintf("[Node] Unable to connect to %v. Permanent failure. %v", url, err)
		log.Println(msg)
		return nil, errors.New(msg)
	}

	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("[Node] Error retrieving stats from %v: %v", n.Api, err)
		log.Println(msg)
		return nil, errors.New(msg)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var stats stats.Stats
	err = json.Unmarshal(body, &stats)
	if err != nil {
		msg := fmt.Sprintf("[Node] Error whilst decoding message for getting stats for node %s", n.Name)
		log.Println(msg)
		return nil, errors.New(msg)
	}

	n.Memory = stats.MemTotalKb()
	n.Disk = int64(stats.DiskTotal())
	n.Stats = stats

	return &n.Stats, nil
}
