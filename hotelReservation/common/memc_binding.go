package common

import (
	"github.com/bradfitz/gomemcache/memcache"
	"hash/crc32"
	"log"
	"net"
	"sort"
	"strconv"
	"time"
)

type MemcachedPool struct {
	addrDomain     string
	consistentHash *ConsistentHash
	monHelper      *MonitoringHelper
}

func NewMemcachedPool(
	monHelper *MonitoringHelper,
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
		monHelper:      monHelper,
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
			Addr:      addr,
			ClientObj: cli,
		})
	}

	sort.Slice(nodes, func(i, j int) bool {
		ipI := ParseIp2Int(nodes[i].Addr)
		ipJ := ParseIp2Int(nodes[j].Addr)
		return ipI < ipJ
	})

	for i, node := range nodes {
		node.Rank = i
	}

	pool.consistentHash.Add(nodes...)
	return pool, nil
}

func (m *MemcachedPool) getConn(key string) (*memcache.Client, int) {
	node := m.consistentHash.GetNode([]byte(key))
	return node.ClientObj.(*memcache.Client), node.Rank
}

func (m *MemcachedPool) Get(key string) (*memcache.Item, error) {
	cli, rank := m.getConn(key)

	startTime := time.Now()
	item, err := cli.Get(key)
	endTime := time.Now()

	reqSize := len(key)

	var respSize int
	if err == nil {
		respSize = len(item.Key) + len(item.Value)
	}

	m.monHelper.submitStoreOpStat3(
		rank,
		startTime,
		endTime,
		m.monHelper.memcStat,
		DbOpRead,
		DbStageRun,
		err,
		reqSize,
		respSize,
		1,
	)

	return item, err
}

func (m *MemcachedPool) Set(item *memcache.Item) error {
	cli, rank := m.getConn(item.Key)

	startTime := time.Now()
	err := cli.Set(item)
	endTime := time.Now()

	reqSize := len(item.Key) + len(item.Value)

	m.monHelper.submitStoreOpStat3(
		rank,
		startTime,
		endTime,
		m.monHelper.memcStat,
		DbOpRead,
		DbStageRun,
		err,
		reqSize,
		0,
		1,
	)
	return err
}
