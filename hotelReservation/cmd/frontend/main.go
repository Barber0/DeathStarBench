package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hotel_reserve/common"
	"hotel_reserve/registry"
	"hotel_reserve/services/frontend"
	"hotel_reserve/tracing"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]string
	json.Unmarshal([]byte(byteValue), &result)
	servIp := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	servPort, _ := strconv.Atoi(result["FrontendPort"])
	if result["Orchestrator"] == "k8s" {
		addrs, _ := net.InterfaceAddrs()
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					servIp = ipnet.IP.String()
				}
			}
		}
		*jaegeraddr = "jaeger:" + strings.Split(result["jaegerAddress"], ":")[1]
		*consuladdr = "consul:" + strings.Split(result["consulAddress"], ":")[1]
	} else {
		servIp = result["FrontendIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()

	fmt.Printf("frontend ip = %s, port = %d\n", servIp, servPort)

	tracer, err := tracing.Init("frontend", *jaegeraddr)
	if err != nil {
		panic(err)
	}

	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	srv := &frontend.Server{
		Registry: registry,
		Tracer:   tracer,
		IpAddr:   servIp,
		Port:     servPort,
		Monitor:  common.NewMonitoringHelper(common.ServiceFrontend, result),
	}
	log.Fatal(srv.Run())
}
