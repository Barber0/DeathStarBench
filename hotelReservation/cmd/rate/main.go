package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"hotel_reserve/common"
	"hotel_reserve/registry"
	"hotel_reserve/services/rate"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"time"
)

func main() {
	paramAgent, err := common.NewParamAgent(common.ServiceRate)
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
	json.Unmarshal(byteValue, &result)

	servPort, _ := strconv.Atoi(result["RatePort"])
	servIp := ""
	rateMongoAddr := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	if result["Orchestrator"] == "k8s" {
		rateMongoAddr = fmt.Sprintf("mongodb-rate:%d", common.MongoPort)
		//rateMongoAddr = "mongodb-rate:" + strings.Split(result["RateMongoAddress"], ":")[1]
		//rateMemcAddr = "memcached-rate:" + strings.Split(result["RateMemcAddress"], ":")[1]
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
		rateMongoAddr = result["RateMongoAddress"]
		servIp = result["RateIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()

	monHelper := common.NewMonitoringHelper(
		common.ServiceRate,
		paramAgent.IpRank,
		result,
	)

	memcIdleConn, _ := strconv.Atoi(common.GetCfgData(common.CfgKeySvrMemcIdleConn, nil))
	memcTimeout, _ := strconv.Atoi(common.GetCfgData(common.CfgKeySvrMemcTimeout, nil))

	memcClient, err := common.NewMemcachedPool(common.ServiceMemcRate, common.MemcachedPort, 10, time.Millisecond*time.Duration(memcTimeout), memcIdleConn)
	if err != nil {
		panic(err)
	}

	mongoSession := initializeDatabase(rateMongoAddr)
	defer mongoSession.Close()
	poolLimit, _ := strconv.Atoi(common.GetCfgData(common.CfgKeySvrDbConn, nil))
	mongoSession.SetPoolLimit(poolLimit)

	fmt.Printf("rate ip = %s, port = %d\n", servIp, servPort)

	//tracer, err := tracing.Init(common.ServiceRate, *jaegeraddr)
	//if err != nil {
	//	panic(err)
	//}

	registryCli, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	srv := &rate.Server{
		//Tracer:       tracer,
		Registry:     registryCli,
		Port:         servPort,
		IpAddr:       servIp,
		MongoSession: mongoSession,
		MemcClient:   memcClient,
		Monitor:      monHelper,
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
