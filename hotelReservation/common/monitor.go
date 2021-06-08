package common

import (
	context2 "context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"os"
	"strings"
	"time"
)

const (
	EnvPodName  = "MY_POD_NAME"
	EnvHostName = "HOSTNAME"

	LabelServiceName = "service"
	LabelPodName     = "pod"
	LabelSrcService  = "src_service"
	LabelSrcPod      = "src_pod"

	PerfLatency = "latency"
)

type MonitoringHelper struct {
	serviceName       string
	podName           string
	metricMap         map[string]prometheus.Gauge
	influxCli         influxdb2.Client
	influxOrg         string
	influxBucket      string
	influxMeasurement string
}

func getPodName() string {
	podName := os.Getenv(EnvPodName)
	if strings.TrimSpace(podName) == "" {
		podName = os.Getenv(EnvHostName)
	}
	return podName
}

var podName = getPodName()

func NewMonitoringHelper(serviceName string, config map[string]string) *MonitoringHelper {
	helper := &MonitoringHelper{
		serviceName:       serviceName,
		podName:           podName,
		metricMap:         make(map[string]prometheus.Gauge),
		influxCli:         influxdb2.NewClient("http://influxdb.autosys:8086", config["InfluxToken"]),
		influxOrg:         config["InfluxOrg"],
		influxBucket:      config["InfluxBucket"],
		influxMeasurement: config["InfluxMeasurement"],
	}

	return helper
}

func (mh *MonitoringHelper) MetricInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context2.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		startTime := time.Now()
		resp, err = handler(ctx, req)
		endTime := time.Now()

		handleDurData := endTime.Sub(startTime).Milliseconds()

		pTag := map[string]string{
			LabelServiceName: mh.serviceName,
			LabelPodName:     mh.podName,
		}

		meta, ok := metadata.FromIncomingContext(ctx)
		if ok {
			pTag[LabelSrcService] = meta.Get(LabelSrcService)[0]
			pTag[LabelSrcPod] = meta.Get(LabelSrcPod)[0]
		}

		writeAPI := mh.influxCli.WriteAPI(mh.influxOrg, mh.influxBucket)
		metricPoint := influxdb2.NewPoint(
			mh.influxMeasurement,
			pTag,
			map[string]interface{}{
				PerfLatency: handleDurData,
			},
			endTime,
		)

		writeAPI.WritePoint(metricPoint)

		return
	}
}
func SenderMetricInterceptor(service string) grpc.UnaryClientInterceptor {
	return func(ctx context2.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		outCtx := metadata.NewOutgoingContext(ctx, metadata.MD{
			LabelSrcService: {service},
			LabelSrcPod:     {podName},
		})

		return invoker(outCtx, method, req, reply, cc, opts...)
	}
}
