package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hotel_reserve/monitor"
	"hotel_reserve/registry"
	"hotel_reserve/services/frontend"
	"hotel_reserve/tracing"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

const ServiceName = "frontend"

func main() {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]string
	json.Unmarshal([]byte(byteValue), &result)
	serv_ip := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	serv_port, _ := strconv.Atoi(result["FrontendPort"])
	if result["Orchestrator"] == "k8s" {
		serv_ip = fmt.Sprintf("%s.%s", monitor.GetPodName(), ServiceName)
		*jaegeraddr = "jaeger:" + strings.Split(result["jaegerAddress"], ":")[1]
		*consuladdr = "consul:" + strings.Split(result["consulAddress"], ":")[1]

	} else {
		serv_ip = result["FrontendIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]

	}
	flag.Parse()

	fmt.Printf("frontend ip = %s, port = %d\n", serv_ip, serv_port)

	tracer, err := tracing.Init(ServiceName, *jaegeraddr)
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
		IpAddr:   serv_ip,
		Port:     serv_port,
	}
	log.Fatal(srv.Run())
}
