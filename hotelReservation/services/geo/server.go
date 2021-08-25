package geo

import (
	// "encoding/json"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"hotel_reserve/common"
	"hotel_reserve/registry"
	pb "hotel_reserve/services/geo/proto"
	"hotel_reserve/tls"
	"strconv"

	"log"
	"net"
	"time"

	"github.com/hailocab/go-geoindex"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	name             = common.ServiceGeo
	maxSearchRadius  = 10
	maxSearchResults = 5
)

// Server implements the geo service
type Server struct {
	index *geoindex.ClusteringIndex
	uuid  string

	Registry *registry.Client
	//Tracer       opentracing.Tracer
	Port         int
	IpAddr       string
	MongoSession *mgo.Session
	Monitor      *common.MonitoringHelper
}

// Run starts the server
func (s *Server) Run() error {
	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
	}

	if s.index == nil {
		s.index = newGeoIndex(s.Monitor, s.MongoSession)
	}

	keepaliveTimeout, _ := strconv.Atoi(common.GetCfgData(common.CfgKeySvrTimeout, nil))

	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Timeout: time.Duration(keepaliveTimeout) * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
		}),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_prometheus.UnaryServerInterceptor,
				s.Monitor.GrpcMetricInterceptor,
				//otgrpc.OpenTracingServerInterceptor(s.Tracer),
			),
		),
	}

	if tlsopt := tls.GetServerOpt(); tlsopt != nil {
		opts = append(opts, tlsopt)
	}

	srv := grpc.NewServer(opts...)
	pb.RegisterGeoServer(srv, s)
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(srv)

	// listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
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

// Nearby returns all hotels within a given distance.
func (s *Server) Nearby(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	// fmt.Printf("In geo Nearby\n")

	var (
		points = s.getNearbyPoints(ctx, float64(req.Lat), float64(req.Lon))
		res    = &pb.Result{}
	)

	// fmt.Printf("geo after getNearbyPoints, len = %d\n", len(points))

	for _, p := range points {
		// fmt.Printf("In geo Nearby return hotelId = %s\n", p.Id())
		res.HotelIds = append(res.HotelIds, p.Id())
	}

	return res, nil
}

func (s *Server) getNearbyPoints(ctx context.Context, lat, lon float64) []geoindex.Point {
	// fmt.Printf("In geo getNearbyPoints, lat = %f, lon = %f\n", lat, lon)

	center := &geoindex.GeoPoint{
		Pid:  "",
		Plat: lat,
		Plon: lon,
	}

	return s.index.KNearest(
		center,
		maxSearchResults,
		geoindex.Km(maxSearchRadius), func(p geoindex.Point) bool {
			return true
		},
	)
}

// newGeoIndex returns a geo index with points loaded
func newGeoIndex(monHelper *common.MonitoringHelper, session *mgo.Session) *geoindex.ClusteringIndex {
	// session, err := mgo.Dial("mongodb-geo")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()

	// fmt.Printf("new geo newGeoIndex\n")

	s := session.Copy()
	defer s.Close()
	c := s.DB("geo-db").C("geo")

	var points []*point
	err := c.Find(bson.M{}).All(&points)
	if err != nil {
		log.Println("Failed get geo data: ", err)
	}

	fmt.Printf("newGeoIndex len(points) = %d\n", len(points))

	// add points to index
	index := geoindex.NewClusteringIndex()
	for _, point := range points {
		index.Add(point)
	}

	return index
}

type point struct {
	Pid  string  `bson:"hotelId"`
	Plat float64 `bson:"lat"`
	Plon float64 `bson:"lon"`
}

// Lat Implement Point interface
func (p *point) Lat() float64 { return p.Plat }
func (p *point) Lon() float64 { return p.Plon }
func (p *point) Id() string   { return p.Pid }
