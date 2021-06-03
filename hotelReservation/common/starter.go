package common

import (
	"github.com/opentracing/opentracing-go"
	"hotel_reserve/registry"
	"hotel_reserve/tracing"
	"net"
	"strconv"
	"strings"
)

func InitService(name string, conf map[string]string) (consulRegistry *registry.Client, tracer opentracing.Tracer, servIp string, servPort int, err error) {
	var (
		jaegeraddr string
		consuladdr string
	)

	servPort, err = strconv.Atoi(conf[HeadToUpper(name)+"Port"])
	if err != nil {
		return
	}

	switch conf["Orchestrator"] {
	case "k8s":
		addresses, _ := net.InterfaceAddrs()
		for _, a := range addresses {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					servIp = ipnet.IP.String()

				}
			}
		}
		jaegeraddr = "jaeger:" + strings.Split(conf["jaegerAddress"], ":")[1]
		consuladdr = "consul:" + strings.Split(conf["consulAddress"], ":")[1]
	case "k8s_istio":
		jaegeraddr = "jaeger:" + strings.Split(conf["jaegerAddress"], ":")[1]
	default:
		jaegeraddr = conf["jaegerAddress"]
		consuladdr = conf["consulAddress"]
	}

	if consuladdr != "" {
		consulRegistry, err = registry.NewClient(consuladdr)
		if err != nil {
			return
		}
	}

	tracer, err = tracing.Init(name, jaegeraddr)
	if err != nil {
		return
	}

	return
}
