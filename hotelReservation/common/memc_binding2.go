package common

import (
	"github.com/bradfitz/gomemcache/memcache"
	"net"
	"strconv"
	"time"
)

type MemcCli struct {
	cli *memcache.Client
}

func NewMemcCli(
	addrDomain string,
	port int,
	timeout time.Duration,
	maxIdleConns int,
) (*memcache.Client, error) {
	addrs, err := net.LookupHost(addrDomain)
	if err != nil {
		return nil, err
	}

	addrsArr := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		addrsArr = append(addrsArr, net.JoinHostPort(addr, strconv.Itoa(port)))
	}
	cli := memcache.New(addrsArr...)
	cli.Timeout = timeout
	cli.MaxIdleConns = maxIdleConns

	return cli, nil
}
