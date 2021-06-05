package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"hotel_reserve/monitor"
	"hotel_reserve/registry"
	"hotel_reserve/services/search"
	"hotel_reserve/tracing"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"strconv"
)

const ServiceName = "search"

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

	servPort, _ := strconv.Atoi(result["SearchPort"])
	if result["Orchestrator"] == "k8s" {
		servIp = fmt.Sprintf("%s.%s", monitor.GetPodName(), ServiceName)

		*jaegeraddr = "jaeger:" + strings.Split(result["jaegerAddress"], ":")[1]
		*consuladdr = "consul:" + strings.Split(result["consulAddress"], ":")[1]

	} else {
		servIp = result["SearchIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]

	}
	flag.Parse()

	fmt.Printf("search ip = %s, port = %d\n", servIp, servPort)

	tracer, err := tracing.Init(ServiceName, *jaegeraddr)
	if err != nil {
		panic(err)
	}

	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	srv := &search.Server{
		Tracer:   tracer,
		Port:     servPort,
		IpAddr:   servIp,
		Registry: registry,
		Monitor:  monitor.NewMonitoringHelper(ServiceName),
	}
	log.Fatal(srv.Run())
}
