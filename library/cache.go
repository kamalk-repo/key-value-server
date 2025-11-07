package library

import "sync"

// --------------- CACHE CODE: START ---------------

// Cache mode enum
type CacheMode int64

const (
	WriteThrough = iota
	WriteBack
)

// Node for linked list datastructure for cache implementation
type CacheItem struct {
	key        int
	value      string
	dirtyBit   int
	prev, next *CacheItem
}

// Wrapper for managing cache operation
type Cache struct {
	capacity int
	cache    map[int]*CacheItem
	head     *CacheItem
	tail     *CacheItem
	mu       sync.Mutex
}

// Create cache with default settings
func NewCache(cap int) *Cache {
	return &Cache{
		capacity: cap,
		cache:    make(map[int]*CacheItem),
	}
}

// Get retrieves a value by key and updates usage order.
func (l *Cache) Get(key int) (string, bool) {
	// l.mu.Lock()
	// defer l.mu.Unlock()

	if node, found := l.cache[key]; found {
		l.moveItemToFront(node)
		return node.value, true
	}
	return "", false
}

// Add or update a key-value pair in the cache.
func (l *Cache) AddOrUpdateCacheItem(key int, value string, kv *KVStore) {
	// l.mu.Lock()
	// defer l.mu.Unlock()

	if node, found := l.cache[key]; found {
		node.value = value
		node.dirtyBit = 1
		l.moveItemToFront(node)
		return
	}

	if len(l.cache) >= l.capacity {
		if kv.mode == WriteBack {
			status, _ := kv.checkIfKeyExists(l.tail.key)
			if status == KeyNotFoundError {
				kv.insertKey(l.tail.key, l.tail.value)
			} else if status == Success {
				kv.updateKey(l.tail.key, l.tail.value)
			}
		}
		delete(l.cache, l.tail.key)
		l.removeItem(l.tail)
	}

	newNode := &CacheItem{key: key, value: value, dirtyBit: 0}
	l.addItemToFront(newNode)
	l.cache[key] = newNode
}

// Delete removes a key-value pair from the cache (if present)
func (l *Cache) Delete(key int) {
	// l.mu.Lock()
	// defer l.mu.Unlock()

	node, found := l.cache[key]
	if !found {
		return // key not present
	}

	// Remove from linked list
	l.removeItem(node)

	// Remove from map
	delete(l.cache, key)
}

// Move a node to the head
func (l *Cache) moveItemToFront(node *CacheItem) {
	if node == l.head {
		return
	}
	l.removeItem(node)
	l.addItemToFront(node)
}

// Removes a node from the linked list
func (l *Cache) removeItem(node *CacheItem) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		l.head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		l.tail = node.prev
	}
}

// Adds a node at the head
func (l *Cache) addItemToFront(node *CacheItem) {
	node.prev = nil
	node.next = l.head
	if l.head != nil {
		l.head.prev = node
	}
	l.head = node
	if l.tail == nil {
		l.tail = node
	}
}

// --------------- CACHE CODE: END ---------------
