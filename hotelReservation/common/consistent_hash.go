package common

import (
	"fmt"
	"sort"
	"sync"
)

type Node struct {
	Addr      string
	ClientObj interface{}
	Rank      int
}

type ConsistentHash struct {
	replicate int
	hashFunc  func(v []byte) uint32
	ring      []*ringNode
	ringBuf   []*ringNode
	valMap    map[Node]struct{}
	keySet    map[string]struct{}
	lock      *sync.RWMutex
}

type ringNode struct {
	hash uint32
	val  Node
}

func NewConsistentHash(rep int, hashFunc func([]byte) uint32) *ConsistentHash {
	ch := &ConsistentHash{
		replicate: rep,
		hashFunc:  hashFunc,
		keySet:    make(map[string]struct{}),
		valMap:    make(map[Node]struct{}),
		ringBuf:   make([]*ringNode, rep),
		lock:      &sync.RWMutex{},
	}
	return ch
}

func (ch *ConsistentHash) Add(nodes ...Node) {
	ch.lock.Lock()
	length := len(ch.ring)
	defer func() {
		if len(ch.ring) != length {
			ch.sortRing()
		}
		ch.lock.Unlock()
	}()
	for _, n := range nodes {
		if _, ok := ch.valMap[n]; ok {
			continue
		}
		ch.keySet[n.Addr] = struct{}{}
		for i := 0; i < ch.replicate; i++ {
			str := fmt.Sprintf("%s#%d", n.Addr, i)
			ch.ringBuf[i] = &ringNode{
				hash: ch.hashFunc([]byte(str)),
				val:  n,
			}
		}
		ch.ring = append(ch.ring, ch.ringBuf...)
		ch.resetRingBuf()
		ch.valMap[n] = struct{}{}
	}
}

func (ch *ConsistentHash) Delete(nodes ...Node) {
	ch.lock.Lock()
	length := len(ch.ring)
	defer func() {
		if length != len(ch.ring) {
			ch.sortRing()
		}
		ch.lock.Unlock()
	}()
	for _, n := range nodes {
		if _, ok := ch.valMap[n]; !ok {
			continue
		}
		for i := 0; i < ch.replicate; i++ {
			length = len(ch.ring)
			objHash := ch.hashFunc([]byte(fmt.Sprintf("%s#%d", n.Addr, i)))
			target := sort.Search(length, func(k int) bool {
				return ch.ring[k].hash >= objHash
			})
			if target < length-1 {
				ch.ring = append(ch.ring[:target], ch.ring[target+1:]...)
			} else {
				ch.ring = ch.ring[:target]
			}
		}
		delete(ch.keySet, n.Addr)
		delete(ch.valMap, n)
	}
}

func (ch *ConsistentHash) In(key string) bool {
	_, ok := ch.keySet[key]
	return ok
}

func (ch *ConsistentHash) GetNode(pkg []byte) Node {
	ch.lock.RLock()
	defer ch.lock.RUnlock()
	length := len(ch.ring)
	objHash := ch.hashFunc(pkg)
	target := sort.Search(length, func(i int) bool {
		return ch.ring[i].hash >= objHash
	})
	return ch.ring[target%length].val
}

func (ch *ConsistentHash) sortRing() {
	sort.Slice(ch.ring, func(i, j int) bool {
		return ch.ring[i].hash < ch.ring[j].hash
	})
}

func (ch *ConsistentHash) resetRingBuf() {
	for i := 0; i < ch.replicate; i++ {
		ch.ringBuf[i] = nil
	}
}

func (ch *ConsistentHash) Show() {
	for _, v := range ch.ring {
		fmt.Println(v.hash, "\t", v.val.Addr)
	}
}
