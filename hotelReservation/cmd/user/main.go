package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"hotel_reserve/common"
	"hotel_reserve/registry"
	"hotel_reserve/services/user"
	"hotel_reserve/tracing"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

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
		userMongoAddr = result["UserMongoAddress"]
		servIp = result["UserIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()

	mongoSession := initializeDatabase(userMongoAddr)
	defer mongoSession.Close()

	fmt.Printf("user ip = %s, port = %d\n", servIp, servPort)

	tracer, err := tracing.Init(common.ServiceUser, *jaegeraddr)
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
		Monitor: common.NewMonitoringHelper(
			common.ServiceUser,
			result,
		),
	}

	wg := new(sync.WaitGroup)
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	wg.Add(1)
	go func() {
		s := <-sigC
		log.Printf("signal[%d] received, shuting down gracefully\n", s)
		srv.Shutdown()
		wg.Done()
	}()

	log.Println(srv.Run())
	wg.Wait()
}
