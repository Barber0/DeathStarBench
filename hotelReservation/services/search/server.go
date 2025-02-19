package search

import (
	// "encoding/json"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"hotel_reserve/common"
	"hotel_reserve/dialer"
	"hotel_reserve/registry"
	geo "hotel_reserve/services/geo/proto"
	rate "hotel_reserve/services/rate/proto"
	"hotel_reserve/tls"
	"strconv"

	// F"io/ioutil"
	"log"
	"net"
	// "os"
	"time"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	pb "hotel_reserve/services/search/proto"
)

const name = common.ServiceSearch

// Server implments the search service
type Server struct {
	//Tracer   opentracing.Tracer
	Registry *registry.Client

	geoClient  geo.GeoClient
	rateClient rate.RateClient

	Port    int
	IpAddr  string
	Monitor *common.MonitoringHelper
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
	pb.RegisterSearchServer(srv, s)
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(srv)

	// init grpc clients
	if err := s.initGeoClient(common.ServiceGeo); err != nil {
		return err
	}
	if err := s.initRateClient(common.ServiceRate); err != nil {
		return err
	}

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

func (s *Server) initGeoClient(name string) error {
	conn, err := dialer.Dial3(
		name,
		s.Monitor,
		dialer.WithBalancer(s.Registry.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.geoClient = geo.NewGeoClient(conn)
	return nil
}

func (s *Server) initRateClient(name string) error {
	conn, err := dialer.Dial3(
		name,
		s.Monitor,
		dialer.WithBalancer(s.Registry.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.rateClient = rate.NewRateClient(conn)
	return nil
}

// Nearby returns ids of nearby hotels ordered by ranking algo
func (s *Server) Nearby(ctx context.Context, req *pb.NearbyRequest) (*pb.SearchResult, error) {
	// find nearby hotels
	// fmt.Printf("in Search Nearby\n")

	// fmt.Printf("nearby lat = %f\n", req.Lat)
	// fmt.Printf("nearby lon = %f\n", req.Lon)

	nearby, err := s.geoClient.Nearby(ctx, &geo.Request{
		Lat: req.Lat,
		Lon: req.Lon,
	})
	if err != nil {
		fmt.Printf("nearby error: %v", err)
		return nil, err
	}

	// for _, hid := range nearby.HotelIds {
	// 	fmt.Printf("get Nearby hotelId = %s\n", hid)
	// }

	// find rates for hotels
	rates, err := s.rateClient.GetRates(ctx, &rate.Request{
		HotelIds: nearby.HotelIds,
		InDate:   req.InDate,
		OutDate:  req.OutDate,
	})
	if err != nil {
		fmt.Printf("rates error: %v", err)
		return nil, err
	}

	// TODO(hw): add simple ranking algo to order hotel ids:
	// * geo distance
	// * price (best discount?)
	// * reviews

	// build the response
	res := new(pb.SearchResult)
	for _, ratePlan := range rates.RatePlans {
		// fmt.Printf("get RatePlan HotelId = %s, Code = %s\n", ratePlan.HotelId, ratePlan.Code)
		res.HotelIds = append(res.HotelIds, ratePlan.HotelId)
	}
	return res, nil
}
