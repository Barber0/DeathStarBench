package common

import (
	"github.com/bradfitz/gomemcache/memcache"
	"hash/crc32"
	"log"
	"net"
	"strconv"
	"time"
)

type MemcachedPool struct {
	addrDomain     string
	consistentHash *ConsistentHash
}

func NewMemcachedPool(
	addrDomain string,
	port int,
	replication int,
	timeout time.Duration,
	maxIdleConns int,
) (*MemcachedPool, error) {
	addrs, err := net.LookupHost(addrDomain)
	if err != nil {
		return nil, err
	}
	pool := &MemcachedPool{
		addrDomain:     addrDomain,
		consistentHash: NewConsistentHash(replication, crc32.ChecksumIEEE),
	}

	nodes := make([]Node, 0, len(addrs))
	for _, addr := range addrs {
		log.Printf("memcached addr: %s, port: %d\n", addr, port)
		cli := memcache.New(net.JoinHostPort(addr, strconv.Itoa(port)))
		cli.Timeout = timeout
		cli.MaxIdleConns = maxIdleConns
		nodes = append(nodes, Node{
			addr,
			cli,
		})
	}

	pool.consistentHash.Add(nodes...)
	return pool, nil
}

func (m *MemcachedPool) getConn(key string) *memcache.Client {
	return m.consistentHash.GetNode([]byte(key)).ClientObj.(*memcache.Client)
}

func (m *MemcachedPool) Get(key string) (*memcache.Item, error) {
	return m.getConn(key).Get(key)
}

func (m *MemcachedPool) Set(item *memcache.Item) error {
	return m.getConn(item.Key).Set(item)
}
