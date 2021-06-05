package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"hotel_reserve/monitor"
	"hotel_reserve/registry"
	"hotel_reserve/services/geo"
	"hotel_reserve/tracing"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const ServiceName = "geo"

func main() {

	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]string
	json.Unmarshal([]byte(byteValue), &result)
	servPort, _ := strconv.Atoi(result["GeoPort"])
	servIp := ""
	geoMongoAddr := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	if result["Orchestrator"] == "k8s" {
		geoMongoAddr = "mongodb-geo:" + strings.Split(result["GeoMongoAddress"], ":")[1]
		servIp = fmt.Sprintf("%s.%s", monitor.GetPodName(), ServiceName)
		*jaegeraddr = "jaeger:" + strings.Split(result["jaegerAddress"], ":")[1]
		*consuladdr = "consul:" + strings.Split(result["consulAddress"], ":")[1]
	} else {
		geoMongoAddr = result["GeoMongoAddress"]
		servIp = result["GeoIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()

	mongoSession := initializeDatabase(geoMongoAddr)
	defer mongoSession.Close()

	fmt.Printf("geo ip = %s, port = %d\n", servIp, servPort)

	tracer, err := tracing.Init(ServiceName, *jaegeraddr)
	if err != nil {
		panic(err)
	}

	registryCli, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	srv := &geo.Server{
		Port:         servPort,
		IpAddr:       servIp,
		Tracer:       tracer,
		Registry:     registryCli,
		MongoSession: mongoSession,
		Monitor:      monitor.NewMonitoringHelper(ServiceName),
	}
	log.Fatal(srv.Run())
}
