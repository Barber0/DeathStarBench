package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"hotel_reserve/common"
	"hotel_reserve/registry"
	"hotel_reserve/services/geo"
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
		geoMongoAddr = fmt.Sprintf("mongodb-geo:%d", common.MongoPort)
		//geoMongoAddr = "mongodb-geo:" + strings.Split(result["GeoMongoAddress"], ":")[1]
		address, _ := net.InterfaceAddrs()
		for _, a := range address {
			if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					servIp = ipNet.IP.String()
				}
			}
		}
		*jaegeraddr = "jaeger:" + strings.Split(result["jaegerAddress"], ":")[1]
		*consuladdr = "consul:" + strings.Split(result["consulAddress"], ":")[1]
	} else {
		geoMongoAddr = result["GeoMongoAddress"]
		servIp = result["GeoIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()

	monHelper := common.NewMonitoringHelper(
		common.ServiceGeo,
		result,
	)

	mongoSession := initializeDatabase(monHelper, geoMongoAddr)
	defer mongoSession.Close()

	poolLimit, _ := strconv.Atoi(common.GetCfgData(common.CfgKeySvrDbConn, nil))
	mongoSession.SetPoolLimit(poolLimit)

	fmt.Printf("geo ip = %s, port = %d\n", servIp, servPort)

	//tracer, err := tracing.Init(common.ServiceGeo, *jaegeraddr)
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

	srv := &geo.Server{
		Port:   servPort,
		IpAddr: servIp,
		//Tracer:       tracer,
		Registry:     registryCli,
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
