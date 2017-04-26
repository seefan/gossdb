package gossdb

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	//	"strings"
	//	"container/heap"
	"github.com/seefan/goerr"
	"sort"
	"strings"
	"sync"
)

type Node struct {
	//host
	ID string
	// The Priority of the item in the queue.
	Priority int
}

type NodeManager struct {
	nodes *PriorityNode
	lock  sync.RWMutex
	root  *Node
}

func NewNodeManager() *NodeManager {
	n := new(NodeManager)
	n.nodes = new(PriorityNode)
	return n
}
func (n *NodeManager) Append(node *Node, weight int) {
	n.lock.Lock()
	defer n.lock.Unlock()
	for i := 0; i < weight; i++ {
		vn := &Node{
			Priority: n.hash(fmt.Sprintf("%s:%d", node.ID, i)),
			ID:       node.ID,
		}
		n.nodes.Append(vn)
	}
	n.nodes.Sort()
	if n.nodes.Len() > 0 {
		n.root = n.nodes.Get(0)
	}
}
func (n *NodeManager) GetNode(key ...string) (string, error) {
	n.lock.RLock()
	defer n.lock.RUnlock()
	if n.root == nil {
		return "", goerr.New("not any node")
	}
	pk := n.hash(strings.Join(key, ":"))

	if n.root.Priority > pk {
		println("pk", pk,n.root.Priority)
		return n.root.ID, nil
	}
	for i := 1; i < n.nodes.Len(); i++ {
		node := n.nodes.Get(i)
		if node.Priority > pk {
			println("pk", pk,node.Priority)
			return node.ID, nil
		}
	}
	println("pk", pk,n.root.Priority)
	return n.root.ID, nil
}
func (n *NodeManager) String() string {
	re := ""
	for i := 0; i < n.nodes.Len(); i++ {
		node := n.nodes.Get(i)
		re += fmt.Sprintf(" %s:%d", node.ID, node.Priority)
	}
	return re
}
func (n *NodeManager) hash(id string) int {
	bytes := md5.Sum([]byte(id))
	return int(binary.LittleEndian.Uint32(bytes[10:14]))
}

type PriorityNode []*Node

func (pn PriorityNode) Len() int { return len(pn) }

func (pn PriorityNode) Less(i, j int) bool {
	return pn[i].Priority < pn[j].Priority
}

func (pn PriorityNode) Swap(i, j int) {
	pn[i], pn[j] = pn[j], pn[i]
}

func (pn *PriorityNode) Append(x *Node) {
	*pn = append(*pn, x)
}

func (pn *PriorityNode) Get(i int) *Node {
	this := *pn
	return this[i]
}
func (pn *PriorityNode) Sort() {
	sort.Sort(PriorityNode(*pn))
}
