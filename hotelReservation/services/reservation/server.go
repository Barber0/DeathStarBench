package reservation

import (
	// "encoding/json"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"hotel_reserve/common"
	"hotel_reserve/registry"
	pb "hotel_reserve/services/reservation/proto"
	"hotel_reserve/tls"

	// "io/ioutil"
	"log"
	"net"
	// "os"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	// "strings"
	"strconv"
)

const name = common.ServiceResv

// Server implements the user service
type Server struct {
	//Tracer       opentracing.Tracer
	Registry     *registry.Client
	Port         int
	IpAddr       string
	MongoSession *mgo.Session
	MemcClient   *memcache.Client
	Monitor      *common.MonitoringHelper
}

// Run starts the server
func (s *Server) Run() error {
	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
	}

	keepaliveTimeout, _ := strconv.Atoi(common.GetCfgData(common.CfgKeySvrTimeout, nil))

	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Timeout: time.Duration(keepaliveTimeout) * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
		}),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor,
			s.Monitor.GrpcMetricInterceptor,
			//otgrpc.OpenTracingServerInterceptor(s.Tracer),
		)),
	}

	if tlsopt := tls.GetServerOpt(); tlsopt != nil {
		opts = append(opts, tlsopt)
	}

	srv := grpc.NewServer(opts...)

	pb.RegisterReservationServer(srv, s)
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(srv)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	err = s.Registry.Register(name, s.IpAddr, s.Port)
	if err != nil {
		return fmt.Errorf("failed register: %v", err)
	}

	return srv.Serve(lis)
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
	s.Registry.Deregister()
	s.Monitor.Close()
}

// MakeReservation makes a reservation based on given information
func (s *Server) MakeReservation(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	res := new(pb.Result)
	res.HotelId = make([]string, 0)

	// session, err := mgo.Dial("mongodb-reservation")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()

	session := s.MongoSession.Copy()
	defer session.Close()

	c := session.DB("reservation-db").C("reservation")
	c1 := session.DB("reservation-db").C("number")

	inDate, _ := time.Parse(
		time.RFC3339,
		req.InDate+"T12:00:00+00:00")

	outDate, _ := time.Parse(
		time.RFC3339,
		req.OutDate+"T12:00:00+00:00")
	hotelId := req.HotelId[0]

	indate := inDate.String()[0:10]

	memcDateNumMap := make(map[string]int)

	for inDate.Before(outDate) {
		// check reservations
		count := 0
		inDate = inDate.AddDate(0, 0, 1)
		outdate := inDate.String()[0:10]

		// first check memc
		memcKey := hotelId + "_" + inDate.String()[0:10] + "_" + outdate
		item, err := s.Monitor.CacheRead(s.MemcClient, memcKey)
		if err == nil {
			// memcached hit
			count, _ = strconv.Atoi(string(item.Value))
			// fmt.Printf("memcached hit %s = %d\n", memc_key, count)
			memcDateNumMap[memcKey] = count + int(req.RoomNumber)

		} else if err == memcache.ErrCacheMiss {
			// memcached miss
			// fmt.Printf("memcached miss\n")
			reserve := make([]reservation, 0)
			err := s.Monitor.DBScan(c, &bson.M{"hotelId": hotelId, "inDate": indate, "outDate": outdate}, &reserve)
			if err != nil {
				panic(err)
			}

			for _, r := range reserve {
				count += r.Number
			}

			memcDateNumMap[memcKey] = count + int(req.RoomNumber)

		} else {
			fmt.Printf("Memmcached error = %s\n", err)
			panic(err)
		}

		// check capacity
		// check memc capacity
		memcCapKey := hotelId + "_cap"
		item, err = s.Monitor.CacheRead(s.MemcClient, memcCapKey)
		hotelCap := 0
		if err == nil {
			// memcached hit
			hotelCap, _ = strconv.Atoi(string(item.Value))
			// fmt.Printf("memcached hit %s = %d\n", memc_cap_key, hotel_cap)
		} else if err == memcache.ErrCacheMiss {
			// memcached miss
			var num number
			err = s.Monitor.DBRead(c1, &bson.M{"hotelId": hotelId}, &num)
			if err != nil {
				panic(err)
			}
			hotelCap = int(num.Number)

			// write to memcache
			s.Monitor.CacheInsert(s.MemcClient, &memcache.Item{Key: memcCapKey, Value: []byte(strconv.Itoa(hotelCap))})
		} else {
			fmt.Printf("Memmcached error = %s\n", err)
			panic(err)
		}

		if count+int(req.RoomNumber) > hotelCap {
			return res, nil
		}
		indate = outdate
	}

	// only update reservation number cache after check succeeds
	for key, val := range memcDateNumMap {
		s.Monitor.CacheUpdate(s.MemcClient, &memcache.Item{Key: key, Value: []byte(strconv.Itoa(val))})
	}

	inDate, _ = time.Parse(
		time.RFC3339,
		req.InDate+"T12:00:00+00:00")

	indate = inDate.String()[0:10]

	for inDate.Before(outDate) {
		inDate = inDate.AddDate(0, 0, 1)
		outdate := inDate.String()[0:10]
		err := s.Monitor.DBInsert(c, &reservation{
			HotelId:      hotelId,
			CustomerName: req.CustomerName,
			InDate:       indate,
			OutDate:      outdate,
			Number:       int(req.RoomNumber),
		})
		if err != nil {
			panic(err)
		}
		indate = outdate
	}

	res.HotelId = append(res.HotelId, hotelId)

	return res, nil
}

// CheckAvailability checks if given information is available
func (s *Server) CheckAvailability(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	res := new(pb.Result)
	res.HotelId = make([]string, 0)

	// session, err := mgo.Dial("mongodb-reservation")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()
	session := s.MongoSession.Copy()
	defer session.Close()

	c := session.DB("reservation-db").C("reservation")
	c1 := session.DB("reservation-db").C("number")

	for _, hotelId := range req.HotelId {
		// fmt.Printf("reservation check hotel %s\n", hotelId)
		inDate, _ := time.Parse(
			time.RFC3339,
			req.InDate+"T12:00:00+00:00")

		outDate, _ := time.Parse(
			time.RFC3339,
			req.OutDate+"T12:00:00+00:00")

		indate := inDate.String()[0:10]

		for inDate.Before(outDate) {
			// check reservations
			count := 0
			inDate = inDate.AddDate(0, 0, 1)
			// fmt.Printf("reservation check date %s\n", inDate.String()[0:10])
			outdate := inDate.String()[0:10]

			// first check memc
			memc_key := hotelId + "_" + inDate.String()[0:10] + "_" + outdate
			item, err := s.Monitor.CacheRead(s.MemcClient, memc_key)

			if err == nil {
				// memcached hit
				count, _ = strconv.Atoi(string(item.Value))
				// fmt.Printf("memcached hit %s = %d\n", memc_key, count)
			} else if err == memcache.ErrCacheMiss {
				// memcached miss

				reserve := make([]reservation, 0)
				err = s.Monitor.DBScan(c, &bson.M{"hotelId": hotelId, "inDate": indate, "outDate": outdate}, &reserve)
				if err != nil {
					panic(err)
				}
				for _, r := range reserve {
					// fmt.Printf("reservation check reservation number = %d\n", hotelId)
					count += r.Number
				}

				// update memcached
				s.Monitor.CacheUpdate(s.MemcClient, &memcache.Item{Key: memc_key, Value: []byte(strconv.Itoa(count))})
			} else {
				fmt.Printf("Memmcached error = %s\n", err)
				panic(err)
			}

			// check capacity
			// check memc capacity
			memc_cap_key := hotelId + "_cap"
			item, err = s.Monitor.CacheRead(s.MemcClient, memc_cap_key)
			hotel_cap := 0

			if err == nil {
				// memcached hit
				hotel_cap, _ = strconv.Atoi(string(item.Value))
				// fmt.Printf("memcached hit %s = %d\n", memc_cap_key, hotel_cap)
			} else if err == memcache.ErrCacheMiss {
				var num number
				err = c1.Find(&bson.M{"hotelId": hotelId}).One(&num)
				if err != nil {
					panic(err)
				}
				hotel_cap = int(num.Number)
				// update memcached
				s.Monitor.CacheInsert(s.MemcClient, &memcache.Item{Key: memc_cap_key, Value: []byte(strconv.Itoa(hotel_cap))})
			} else {
				fmt.Printf("Memmcached error = %s\n", err)
				panic(err)
			}

			if count+int(req.RoomNumber) > hotel_cap {
				break
			}
			indate = outdate

			if inDate.Equal(outDate) {
				res.HotelId = append(res.HotelId, hotelId)
			}
		}
	}

	return res, nil
}

type reservation struct {
	HotelId      string `bson:"hotelId"`
	CustomerName string `bson:"customerName"`
	InDate       string `bson:"inDate"`
	OutDate      string `bson:"outDate"`
	Number       int    `bson:"number"`
}

type number struct {
	HotelId string `bson:"hotelId"`
	Number  int    `bson:"numberOfRoom"`
}
