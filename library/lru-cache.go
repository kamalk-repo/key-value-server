package library

type Node struct {
	key      int
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
	hash  map[int]*Node
}

func NewQueue(capacity int) Queue {
	head := &Node{}
	tail := &Node{}
	head.right = tail
	tail.left = head

	return Queue{head: head, tail: tail, capacity: capacity}
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{queue: NewQueue(capacity), hash: make(map[int]*Node)}
}

func (lc *LRUCache) CheckKey(key int) (*Node, bool) {
	node := &Node{}
	keyExists := false
	if val, ok := lc.hash[key]; ok {
		keyExists = true
		if lc.queue.head.right.key != val.key {
			// Put found key in the front
			lc.Remove(val)
			lc.Add(node)
		}
		node = val
	}

	return node, keyExists
}

func (lc *LRUCache) UpdateKey(key int, keyValue string) (*Node, bool) {
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
	left := n.left
	right := n.right

	left.right = right
	right.left = left
	lc.queue.length -= 1
	delete(lc.hash, n.key)
	return n
}

func (lc *LRUCache) Add(n *Node) {
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
