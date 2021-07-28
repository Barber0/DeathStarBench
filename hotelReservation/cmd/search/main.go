package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"hotel_reserve/common"
	"hotel_reserve/registry"
	"hotel_reserve/services/search"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"strconv"
)

func main() {
	paramAgent, err := common.NewParamAgent(common.ServiceSearch)
	if err != nil {
		panic(err)
	}
	fmt.Println("param agent ip rank: ", paramAgent.IpRank)

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

	servPort, _ := strconv.Atoi(result["SearchPort"])
	if result["Orchestrator"] == "k8s" {
		address, _ := net.InterfaceAddrs()
		for _, a := range address {
			if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					servIp = ipNet.IP.String()
				}
			}
		}
		*jaegeraddr = "jaeger:" + strings.Split(result["jaegerAddress"], ":")[1]
		*consuladdr = "consul:" + strings.Split(result["consulAddress"], ":")[1]

	} else {
		servIp = result["SearchIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]

	}
	flag.Parse()

	fmt.Printf("search ip = %s, port = %d\n", servIp, servPort)

	//tracer, err := tracing.Init(common.ServiceSearch, *jaegeraddr)
	//if err != nil {
	//	panic(err)
	//}

	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	srv := &search.Server{
		//Tracer:   tracer,
		Port:     servPort,
		IpAddr:   servIp,
		Registry: registry,
		Monitor: common.NewMonitoringHelper(
			common.ServiceSearch,
			result,
		),
	}

	sigC := make(chan os.Signal)
	signal.Notify(sigC, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Println(srv.Run())
	}()

	sig := <-sigC
	log.Printf("receive signal: %v\n", sig)
	srv.Shutdown()
	log.Println("service shutdown")
}
