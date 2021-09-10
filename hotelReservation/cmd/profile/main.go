package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"hotel_reserve/common"
	"hotel_reserve/registry"
	"hotel_reserve/services/profile"
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
	paramAgent, err := common.NewParamAgent(common.ServiceProfile)
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

	servPort, _ := strconv.Atoi(result["ProfilePort"])
	servIp := ""
	profileMongoAddr := ""

	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	if result["Orchestrator"] == "k8s" {
		profileMongoAddr = fmt.Sprintf("mongodb-profile:%d", common.MongoPort)
		//profileMongoAddr = "mongodb-profile:" + strings.Split(result["ProfileMongoAddress"], ":")[1]
		//profileMemcAddr = "memcached-profile:" + strings.Split(result["ProfileMemcAddress"], ":")[1]
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
		profileMongoAddr = result["ProfileMongoAddress"]
		servIp = result["ProfileIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()

	monHelper := common.NewMonitoringHelper(
		common.ServiceProfile,
		paramAgent.IpRank,
		result,
	)

	mongoSession := initializeDatabase(profileMongoAddr)
	defer mongoSession.Close()

	poolLimit, _ := strconv.Atoi(common.GetCfgData(common.CfgKeySvrDbConn, nil))
	mongoSession.SetPoolLimit(poolLimit)

	memcIdleConn, _ := strconv.Atoi(common.GetCfgData(common.CfgKeySvrMemcIdleConn, nil))
	memcTimeout, _ := strconv.Atoi(common.GetCfgData(common.CfgKeySvrMemcTimeout, nil))

	memcClient, err := common.NewMemcachedPool(common.ServiceMemcProfile, common.MemcachedPort, 10, time.Millisecond*time.Duration(memcTimeout), memcIdleConn)
	if err != nil {
		panic(err)
	}

	fmt.Printf("profile ip = %s, port = %d\n", servIp, servPort)

	//registryCli, err := registry.NewClient(*consuladdr)
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

	srv := profile.Server{
		//Tracer: tracer,
		// Port:     *port,
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
