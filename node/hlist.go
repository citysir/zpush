package main

// HlistNode is an node of a linked hlist.
type HlistNode struct {
	next  *HlistNode
	pprev **HlistNode

	Conn *Connection
}

// Hlist represents a doubly linked hlist.
// The zero value for Hlist is an empty Hlist ready to use.
type Hlist struct {
	first *HlistNode // sentinel hlist first
	len   int        // current hlist length excluding (this) sentinel node
}

// Next returns the next hlist node or nil.
func (e *HlistNode) Next() *HlistNode {
	return e.next
}

// Len returns the number of elements of hlist l.
// The complexity is O(1).
func (l *Hlist) Len() int { return l.len }

// Front returns the first node of hlist l or nil
func (l *Hlist) Front() *HlistNode {
	return l.first
}

// PushFront inserts a new node n with conn at the front of hlist l and returns n.
func (l *Hlist) PushFront(conn *Connection) *HlistNode {
	first := l.first
	n := &HlistNode{Conn: conn}
	n.next = first
	if first != nil {
		first.pprev = &n.next
	}
	l.first = n
	n.pprev = &l.first
	l.len++
	return n
}

// Remove removes e from l if e is an node of hlist l.
// It returns the node value n.Conn.
func (l *Hlist) Remove(n *HlistNode) *Connection {
	next := n.next
	pprev := npprev
	*pprev = next
	if next != nil {
		next.pprev = pprev
	}
	l.len--
	n.next = nil  // avoid memory leak
	n.pprev = nil // avoid memory leak
	return n.Conn
}
