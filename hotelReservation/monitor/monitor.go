package monitor

import (
	context2 "context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"log"
	"os"
	"strings"
	"time"
)

const (
	MetricName = "service_handle_millisecond_duration"

	EnvPodName  = "MY_POD_NAME"
	EnvHostName = "HOSTNAME"

	LabelServiceName = "service"
	LabelPodName     = "pod"
	LabelMethodName  = "method"
)

type MonitoringHelper struct {
	serviceName string
	podName     string
	metricMap   map[string]prometheus.Gauge
}

func NewMonitoringHelper(serviceName string) *MonitoringHelper {
	podName := os.Getenv(EnvPodName)
	if strings.TrimSpace(podName) == "" {
		podName = os.Getenv(EnvHostName)
	}

	helper := &MonitoringHelper{
		serviceName: serviceName,
		podName:     podName,
		metricMap:   make(map[string]prometheus.Gauge),
	}

	return helper
}

func (mh *MonitoringHelper) getServerMetric(methodName string) (metric prometheus.Gauge) {
	defer func() {
		if pa := recover(); pa != nil {
			log.Printf("register failed: %v\n", pa)
			metric = nil
		}
	}()

	var ok bool

	metric, ok = mh.metricMap[methodName]
	if ok {
		return metric
	}

	metric = promauto.NewGauge(prometheus.GaugeOpts{
		Name: MetricName,
		ConstLabels: map[string]string{
			LabelServiceName: mh.serviceName,
			LabelPodName:     mh.podName,
			LabelMethodName:  methodName,
		},
	})

	mh.metricMap[methodName] = metric

	return metric
}

func (mh *MonitoringHelper) MetricInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context2.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		startTime := time.Now()
		resp, err = handler(ctx, req)
		endTime := time.Now()

		handleDurData := endTime.Sub(startTime).Milliseconds()

		if metric := mh.getServerMetric(info.FullMethod); metric != nil {
			metric.Set(float64(handleDurData))
		}
		return
	}
}
