package user

import (
	"crypto/sha256"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"hotel_reserve/common"
	"hotel_reserve/registry"
	"hotel_reserve/tls"
	"strconv"

	// "encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	pb "hotel_reserve/services/user/proto"
	// "io/ioutil"
	"log"
	"net"
	// "os"
	"time"
)

const name = common.ServiceUser

// Server implements the user service
type Server struct {
	//Tracer       opentracing.Tracer
	Registry     *registry.Client

	//users map[string]string
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

	//if s.users == nil {
	//	s.users = loadUsers(s.Monitor, s.MongoSession)
	//}

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
	pb.RegisterUserServer(srv, s)
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(srv)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// // register the service
	// jsonFile, err := os.Open("config.json")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// defer jsonFile.Close()

	// byteValue, _ := ioutil.ReadAll(jsonFile)

	// var result map[string]string
	// json.Unmarshal([]byte(byteValue), &result)

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

// CheckUser returns whether the username and password are correct.
func (s *Server) CheckUser(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	res := new(pb.Result)

	sum := sha256.Sum256([]byte(req.Password))
	pass := fmt.Sprintf("%x", sum)

	sess := s.MongoSession.Copy()
	defer sess.Close()

	c := sess.DB("user-db").C("user")

	user := User{}
	err := s.Monitor.DBRead(c, bson.M{"username": req.Username}, &user)
	if err != nil {
		return nil, err
	}

	res.Correct = pass == user.Password

	return res, nil
}

// loadUsers loads hotel users from mongodb.
func loadUsers(monHelper *common.MonitoringHelper, session *mgo.Session) map[string]string {
	s := session.Copy()
	defer s.Close()
	c := s.DB("user-db").C("user")

	// unmarshal json profiles
	var users []User
	err := monHelper.DBScan(c, bson.M{}, &users)
	if err != nil {
		log.Println("Failed get users data: ", err)
	}

	res := make(map[string]string)
	for _, user := range users {
		res[user.Username] = user.Password
	}

	fmt.Printf("Done load users\n")

	return res
}

type User struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
}
