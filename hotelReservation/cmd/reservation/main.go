package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"hotel_reserve/common"
	"hotel_reserve/registry"
	"hotel_reserve/services/reservation"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	"time"
)

func main() {
	paramAgent, err := common.NewParamAgent(common.ServiceResv)
	if err != nil {
		panic(err)
	}
	log.Printf("init param agent, ip rank: %d\n", paramAgent.IpRank)

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
	var reserveMongoAddr string
	reserveMemcAddr := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	if result["Orchestrator"] == "k8s" {
		reserveMongoAddr = fmt.Sprintf("mongodb-reserve:%d", common.MongoPort)
		reserveMemcAddr = fmt.Sprintf("memcached-reserve:%d", common.MemcachedPort)
		//reserveMemcAddr = "memcached-reserve:" + strings.Split(result["ReserveMemcAddress"], ":")[1]
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

	monHelper := common.NewMonitoringHelper(
		common.ServiceResv,
		paramAgent.IpRank,
		result,
	)

	mongoSession := initializeDatabase(monHelper, reserveMongoAddr)
	defer mongoSession.Close()

	poolLimit, _ := strconv.Atoi(common.GetCfgData(common.CfgKeySvrDbConn, nil))
	mongoSession.SetPoolLimit(poolLimit)

	memcIdleConn, _ := strconv.Atoi(common.GetCfgData(common.CfgKeySvrMemcIdleConn, nil))
	memcTimeout, _ := strconv.Atoi(common.GetCfgData(common.CfgKeySvrMemcTimeout, nil))

	fmt.Printf("reservation memc addr port = %s\n", result["ReserveMemcAddress"])
	memcClient := memcache.New(reserveMemcAddr)
	memcClient.Timeout = time.Second * time.Duration(memcTimeout)
	memcClient.MaxIdleConns = memcIdleConn

	fmt.Printf("reservation ip = %s, port = %d\n", servIp, servPort)

	//tracer, err := tracing.Init(common.ServiceResv, *jaegeraddr)
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

	srv := &reservation.Server{
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
