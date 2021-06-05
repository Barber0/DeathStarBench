package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"hotel_reserve/monitor"
	"hotel_reserve/registry"
	"hotel_reserve/services/user"
	"hotel_reserve/tracing"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const ServiceName = "user"

func main() {
	// initializeDatabase()
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]string
	json.Unmarshal(byteValue, &result)

	servPort, _ := strconv.Atoi(result["UserPort"])
	servIp := ""
	userMongoAddr := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	if result["Orchestrator"] == "k8s" {
		userMongoAddr = "mongodb-user:" + strings.Split(result["UserMongoAddress"], ":")[1]

		servIp = fmt.Sprintf("%s.%s", monitor.GetPodName(), ServiceName)

		*jaegeraddr = "jaeger:" + strings.Split(result["jaegerAddress"], ":")[1]
		*consuladdr = "consul:" + strings.Split(result["consulAddress"], ":")[1]
	} else {
		userMongoAddr = result["UserMongoAddress"]
		servIp = result["UserIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()

	mongoSession := initializeDatabase(userMongoAddr)
	defer mongoSession.Close()

	fmt.Printf("user ip = %s, port = %d\n", servIp, servPort)

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

	srv := &user.Server{
		Tracer:       tracer,
		Registry:     registryCli,
		Port:         servPort,
		IpAddr:       servIp,
		MongoSession: mongoSession,
		Monitor:      monitor.NewMonitoringHelper(ServiceName),
	}
	log.Fatal(srv.Run())
}
