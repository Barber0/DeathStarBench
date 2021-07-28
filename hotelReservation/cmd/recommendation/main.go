package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"hotel_reserve/common"
	"hotel_reserve/registry"
	"hotel_reserve/services/recommendation"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	paramAgent, err := common.NewParamAgent(common.ServiceReco)
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

	servPort, _ := strconv.Atoi(result["RecommendPort"])
	servIp := ""
	recommendationMongoAddr := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	if result["Orchestrator"] == "k8s" {
		recommendationMongoAddr = fmt.Sprintf("mongodb-recommendation:%d", common.MongoPort)
		//recommendationMongoAddr = "mongodb-recommendation:" + strings.Split(result["RecommendMongoAddress"], ":")[1]
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
		recommendationMongoAddr = result["RecommendMongoAddress"]
		servIp = result["RecommendIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()

	monHelper := common.NewMonitoringHelper(
		common.ServiceReco,
		result,
	)

	mongoSession := initializeDatabase(monHelper, recommendationMongoAddr)
	defer mongoSession.Close()

	poolLimit, _ := strconv.Atoi(common.GetCfgData(common.CfgKeySvrDbConn, nil))
	mongoSession.SetPoolLimit(poolLimit)

	fmt.Printf("recommendation ip = %s, port = %d\n", servIp, servPort)

	//tracer, err := tracing.Init(common.ServiceReco, *jaegeraddr)
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

	srv := &recommendation.Server{
		//Tracer:       tracer,
		Registry:     registryCli,
		Port:         servPort,
		IpAddr:       servIp,
		MongoSession: mongoSession,
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
