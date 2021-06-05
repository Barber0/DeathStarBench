package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"hotel_reserve/monitor"
	"hotel_reserve/registry"
	"hotel_reserve/services/profile"
	"hotel_reserve/tracing"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
	"time"
)

const ServiceName = "profile"

func main() {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]string
	json.Unmarshal([]byte(byteValue), &result)

	servPort, _ := strconv.Atoi(result["ProfilePort"])
	servIp := ""
	profileMongoAddr := ""
	profileMemcAddr := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	if result["Orchestrator"] == "k8s" {
		profileMongoAddr = "mongodb-profile:" + strings.Split(result["ProfileMongoAddress"], ":")[1]
		profileMemcAddr = "memcached-profile:" + strings.Split(result["ProfileMemcAddress"], ":")[1]

		servIp = fmt.Sprintf("%s.%s", monitor.GetPodName(), ServiceName)

		*jaegeraddr = "jaeger:" + strings.Split(result["jaegerAddress"], ":")[1]
		*consuladdr = "consul:" + strings.Split(result["consulAddress"], ":")[1]
	} else {
		profileMongoAddr = result["ProfileMongoAddress"]
		profileMemcAddr = result["ProfileMemcAddress"]
		servIp = result["ProfileIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()

	mongoSession := initializeDatabase(profileMongoAddr)
	defer mongoSession.Close()

	fmt.Printf("profile memc addr port = %s\n", profileMemcAddr)
	memcClient := memcache.New(profileMemcAddr)
	memcClient.Timeout = time.Second * 2
	memcClient.MaxIdleConns = 512

	fmt.Printf("profile ip = %s, port = %d\n", servIp, servPort)

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

	srv := profile.Server{
		Tracer: tracer,
		// Port:     *port,
		Registry:     registryCli,
		Port:         servPort,
		IpAddr:       servIp,
		MongoSession: mongoSession,
		MemcClient:   memcClient,
		Monitor:      monitor.NewMonitoringHelper(ServiceName),
	}
	log.Fatal(srv.Run())
}
