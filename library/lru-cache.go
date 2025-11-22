package library

import "sync"

type Node struct {
	key      string
	keyValue string
	left     *Node
	right    *Node
}

type Queue struct {
	head     *Node
	tail     *Node
	length   int
	capacity int
}

type LRUCache struct {
	queue Queue
	mu    sync.RWMutex
	hash  map[string]*Node
}

func NewQueue(capacity int) Queue {
	head := &Node{}
	tail := &Node{}
	head.right = tail
	tail.left = head

	return Queue{head: head, tail: tail, capacity: capacity}
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{queue: NewQueue(capacity), hash: make(map[string]*Node)}
}

func (lc *LRUCache) CheckKey(key string) (*Node, bool) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	if val, ok := lc.hash[key]; ok {
		return val, true
	}

	return &Node{}, false
}

func (lc *LRUCache) UpdateKey(key string, keyValue string) (*Node, bool) {
	node := &Node{}
	keyExists := false

	if val, ok := lc.hash[key]; ok {
		keyExists = true
		// Update key value
		val.keyValue = keyValue

		if lc.queue.head.right.key != val.key {
			// Put updated key in the front
			lc.Remove(val)
			lc.Add(val)
		}
		node = val
	}

	return node, keyExists
}

func (lc *LRUCache) Remove(n *Node) *Node {
	// lc.mu.Lock()
	// defer lc.mu.Unlock()
	left := n.left
	right := n.right

	left.right = right
	right.left = left
	lc.queue.length -= 1
	delete(lc.hash, n.key)
	return n
}

func (lc *LRUCache) Add(n *Node) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	tmp := lc.queue.head.right
	lc.queue.head.right = n
	n.left = lc.queue.head
	n.right = tmp
	tmp.left = n
	lc.hash[n.key] = n

	lc.queue.length += 1
	if lc.queue.length > lc.queue.capacity {
		lc.Remove(lc.queue.tail.left)
	}
}
