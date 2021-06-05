package monitor

import (
	context2 "context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/push"
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

func GetPodName() string {
	podName := os.Getenv(EnvPodName)
	if strings.TrimSpace(podName) == "" {
		podName = os.Getenv(EnvHostName)
	}
	return podName
}

type MonitoringHelper struct {
	serviceName string
	podName     string
	metricMap   map[string]prometheus.Gauge
}

func NewMonitoringHelper(serviceName string) *MonitoringHelper {
	podName := GetPodName()

	helper := &MonitoringHelper{
		serviceName: serviceName,
		podName:     podName,
		metricMap:   make(map[string]prometheus.Gauge),
	}

	return helper
}

func (mh *MonitoringHelper) getServerMetric(methodName string) prometheus.Gauge {
	metric, ok := mh.metricMap[methodName]
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

	promRegistry := prometheus.NewRegistry()
	if err := promRegistry.Register(metric); err != nil {
		log.Printf("register metric[%s] failed: %v\n", methodName, err)
		return nil
	}

	pusher := push.New(
		"prom-pushgateway:9091",
		fmt.Sprintf("server_handle_duration_%s_%s",
			mh.serviceName,
			methodName,
		),
	)

	pusher.Gatherer(promRegistry)
	pusher.Add()
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
