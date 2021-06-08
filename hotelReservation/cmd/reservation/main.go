package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"hotel_reserve/common"
	"hotel_reserve/registry"
	"hotel_reserve/services/reservation"
	"hotel_reserve/tracing"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	"time"
)

func main() {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]string
	json.Unmarshal(byteValue, &result)
	servPort, _ := strconv.Atoi(result["ReservePort"])
	servIp := ""
	reserveMongoAddr := ""
	reserveMemcAddr := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	if result["Orchestrator"] == "k8s" {
		reserveMongoAddr = "mongodb-reserve:" + strings.Split(result["ReserveMongoAddress"], ":")[1]
		reserveMemcAddr = "memcached-reserve:" + strings.Split(result["ReserveMemcAddress"], ":")[1]
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
		reserveMongoAddr = result["ReserveMongoAddress"]
		reserveMemcAddr = result["ReserveMemcAddress"]
		servIp = result["ReserveIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()

	mongoSession := initializeDatabase(reserveMongoAddr)
	defer mongoSession.Close()

	fmt.Printf("reservation memc addr port = %s\n", result["ReserveMemcAddress"])
	memcClient := memcache.New(reserveMemcAddr)
	memcClient.Timeout = time.Second * 2
	memcClient.MaxIdleConns = 512

	fmt.Printf("reservation ip = %s, port = %d\n", servIp, servPort)

	tracer, err := tracing.Init(common.ServiceResv, *jaegeraddr)
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

	srv := &reservation.Server{
		Tracer:       tracer,
		Registry:     registryCli,
		Port:         servPort,
		IpAddr:       servIp,
		MongoSession: mongoSession,
		MemcClient:   memcClient,
		Monitor: common.NewMonitoringHelper(
			common.ServiceResv,
			result,
		),
	}
	log.Fatal(srv.Run())
}
