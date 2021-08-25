package common

import (
	context2 "context"
	"errors"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

const (
	EnvPodName  = "MY_POD_NAME"
	EnvHostName = "HOSTNAME"

	LabelServiceName = "service"
	LabelPodName     = "pod"
	LabelMethod      = "method"
	LabelSrcService  = "src_service"
	LabelSrcPod      = "src_pod"
	LabelFailed      = "failed"
	LabelDbOpType    = "op"
	LabelDbStage     = "stage"
	LabelCrashed     = "crash"
	LabelEpoch       = "epoch"

	DbStageLoad = "load"
	DbStageRun  = "run"

	DbOpRead   = "read"
	DbOpScan   = "scan"
	DbOpInsert = "insert"
	DbOpUpdate = "update"

	DummyTagVal = "true"

	MetricReqSize  = "req_size"
	MetricRespSize = "resp_size"
	MetricRespLen  = "resp_len"
	PerfLatency    = "latency"

	DummySrcPodWrk = "wrk"
	DummySrcSvcWrk = DummySrcPodWrk

	ReqHeaderEpoch = "epoch"

	IntSize = 8
)

type MonitoringHelper struct {
	serviceName  string
	podName      string
	metricMap    map[string]prometheus.Gauge
	influxCli    influxdb2.Client
	writeAPI     api.WriteAPI
	influxOrg    string
	influxBucket string
	serviceStat  string
	mgoStat      string
	memcStat     string
	benchEpoch   string
}

type (
	dbStatFunc1 func(string, func() error) error
	dbStatFunc2 func(string, func() (int, error)) (int, error)

	cacheStatFunc1 func(string, func() error) error
	cacheStatFunc2 func(string, func() (*memcache.Item, error)) (*memcache.Item, error)
)

func getPodName() string {
	podName := os.Getenv(EnvPodName)
	if strings.TrimSpace(podName) == "" {
		podName = os.Getenv(EnvHostName)
	}
	return podName
}

var podName = getPodName()

func NewMonitoringHelper(serviceName string, config map[string]string) *MonitoringHelper {

	influxBatchSize, _ := strconv.Atoi(GetCfgData(CfgKeyInfluxBatchSize, config))
	influxFlushInterval, _ := strconv.Atoi(GetCfgData(CfgKeyInfluxFlushInterval, config))
	opt := influxdb2.DefaultOptions().
		SetBatchSize(uint(influxBatchSize)).
		SetFlushInterval(uint(influxFlushInterval))
	cli := influxdb2.NewClientWithOptions("http://influxdb.autosys:8086", GetCfgData(CfgKeyInfluxToken, config), opt)

	helper := &MonitoringHelper{
		serviceName: serviceName,
		podName:     podName,
		metricMap:   make(map[string]prometheus.Gauge),
		influxCli:   cli,
		writeAPI: cli.WriteAPI(
			GetCfgData(CfgKeyInfluxOrg, config),
			GetCfgData(CfgKeyInfluxBucket, config),
		),
		serviceStat: GetCfgData(CfgKeyServiceStat, config),
		mgoStat:     GetCfgData(CfgKeyMgoStat, config),
		memcStat:    GetCfgData(CfgKeyMemcStat, config),
	}

	return helper
}

func (mh *MonitoringHelper) getCtxData(m map[string]string, md metadata.MD, keys ...string) {
	for _, key := range keys {
		if dataArr := md.Get(key); len(dataArr) > 0 {
			m[key] = dataArr[0]
		}
	}
}

func GetCfgData(key string, config map[string]string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	if config == nil {
		return ""
	}
	return config[key]
}

func (mh *MonitoringHelper) executeHttpHandler(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) (failed bool) {
	defer func() {
		if pa := recover(); pa != nil {
			errInfo := fmt.Sprintf("[ERR] http metric interceptor panic: %v\n", pa)
			http.Error(w, errInfo, http.StatusInternalServerError)
			log.Println(errInfo)
			fmt.Println(string(debug.Stack()))
			failed = true
		}
	}()
	next.ServeHTTP(w, r)
	return
}

func (mh *MonitoringHelper) HttpMetricInterceptor(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mh.benchEpoch = r.Header.Get(ReqHeaderEpoch)

		startTime := time.Now()
		reqFailed := mh.executeHttpHandler(w, r, next)
		endTime := time.Now()

		handleDurData := endTime.Sub(startTime).Microseconds()

		pTag := map[string]string{
			LabelServiceName: mh.serviceName,
			LabelPodName:     mh.podName,
			LabelMethod:      r.URL.Path,
			LabelSrcPod:      DummySrcPodWrk,
			LabelSrcService:  DummySrcSvcWrk,
			LabelEpoch:       mh.benchEpoch,
		}

		if reqFailed {
			pTag[LabelFailed] = DummyTagVal
		}

		metricPoint := influxdb2.NewPoint(
			mh.serviceStat,
			pTag,
			map[string]interface{}{
				PerfLatency: handleDurData,
			},
			startTime,
		)

		mh.writeAPI.WritePoint(metricPoint)
	}
}

func (mh *MonitoringHelper) executeHandler(ctx context2.Context, req interface{}, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if pa := recover(); pa != nil {
			resp = nil

			switch pa.(type) {
			case string:
				err = errors.New(pa.(string))
			case error:
				err = pa.(error)
			default:
				err = fmt.Errorf("%v", pa)
			}
		}
	}()
	resp, err = handler(ctx, req)
	return
}

func (mh *MonitoringHelper) GrpcMetricInterceptor(ctx context2.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	startTime := time.Now()
	resp, err = mh.executeHandler(ctx, req, handler)
	endTime := time.Now()

	handleDurData := endTime.Sub(startTime).Microseconds()

	pTag := map[string]string{
		LabelServiceName: mh.serviceName,
		LabelPodName:     mh.podName,
		LabelMethod:      info.FullMethod,
	}

	if err != nil {
		pTag[LabelFailed] = DummyTagVal
	}

	meta, ok := metadata.FromIncomingContext(ctx)
	if ok {
		mh.getCtxData(
			pTag,
			meta,
			LabelSrcService,
			LabelSrcPod,
			LabelEpoch,
		)
		if epochVal, ok := pTag[LabelEpoch]; ok {
			mh.benchEpoch = epochVal
		}
	}

	metricPoint := influxdb2.NewPoint(
		mh.serviceStat,
		pTag,
		map[string]interface{}{
			PerfLatency: handleDurData,
		},
		startTime,
	)

	mh.writeAPI.WritePoint(metricPoint)
	return
}

func (mh *MonitoringHelper) Close() {
	mh.writeAPI.Flush()
	mh.influxCli.Close()
}

func (mh *MonitoringHelper) SenderMetricInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context2.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		outCtx := metadata.NewOutgoingContext(ctx, metadata.MD{
			LabelSrcService: {mh.serviceName},
			LabelSrcPod:     {mh.podName},
			LabelEpoch:      {mh.benchEpoch},
		})
		return invoker(outCtx, method, req, reply, cc, opts...)
	}
}

func (mh *MonitoringHelper) submitStoreOpStat2(
	startTime,
	endTime time.Time,
	table,
	opType,
	stage string,
	err error,
	reqSize,
	respSize int,
	respLen int,
) {
	pTag := map[string]string{
		LabelSrcService: mh.serviceName,
		LabelSrcPod:     mh.podName,
		LabelDbOpType:   opType,
		LabelDbStage:    stage,
		LabelEpoch:      mh.benchEpoch,
	}

	pData := map[string]interface{}{
		MetricReqSize:  reqSize,
		MetricRespSize: respSize,
		MetricRespLen:  respLen,

		PerfLatency: endTime.Sub(startTime).Microseconds(),
	}

	if err != nil {
		pTag[LabelFailed] = DummyTagVal
	}

	metricPoint := influxdb2.NewPoint(
		table,
		pTag,
		pData,
		startTime,
	)

	mh.writeAPI.WritePoint(metricPoint)
}

func (mh *MonitoringHelper) submitStoreOpStat(
	startTime,
	endTime time.Time,
	table,
	opType,
	stage string,
	err error,
) {
	pTag := map[string]string{
		LabelSrcService: mh.serviceName,
		LabelSrcPod:     mh.podName,
		LabelDbOpType:   opType,
		LabelDbStage:    stage,
		LabelEpoch:      mh.benchEpoch,
	}

	pData := map[string]interface{}{
		PerfLatency: endTime.Sub(startTime).Microseconds(),
	}

	if err != nil {
		pTag[LabelFailed] = DummyTagVal
	}

	metricPoint := influxdb2.NewPoint(
		table,
		pTag,
		pData,
		startTime,
	)

	mh.writeAPI.WritePoint(metricPoint)
}

func (mh *MonitoringHelper) CacheStatTool(stage string) (cacheStatFunc1, cacheStatFunc2) {
	return func(op string, f func() error) error {
			startTime := time.Now()
			err := f()
			endTime := time.Now()
			mh.submitStoreOpStat(startTime, endTime, mh.memcStat, op, stage, err)
			return err
		}, func(op string, f func() (*memcache.Item, error)) (*memcache.Item, error) {
			startTime := time.Now()
			it, err := f()
			endTime := time.Now()
			mh.submitStoreOpStat(startTime, endTime, mh.memcStat, op, stage, err)
			return it, err
		}
}

func (mh *MonitoringHelper) DBStatTool(stage string) (dbStatFunc1, dbStatFunc2) {
	return func(opType string, callback func() error) error {
			startTime := time.Now()
			err := callback()
			endTime := time.Now()
			mh.submitStoreOpStat(startTime, endTime, mh.mgoStat, opType, stage, err)
			return err
		}, func(opType string, callback func() (int, error)) (int, error) {
			startTime := time.Now()
			count, err := callback()
			endTime := time.Now()
			mh.submitStoreOpStat(startTime, endTime, mh.mgoStat, opType, stage, err)
			return count, err
		}
}

func (mh *MonitoringHelper) BsonSize(v interface{}) (int, error) {
	bsonBts, err := bson.Marshal(v)
	if err != nil {
		return 0, err
	}
	return len(bsonBts), nil
}

func (mh *MonitoringHelper) BsonSizeLen(v interface{}) (int, int, error) {
	bts, err := bson.Marshal(v)
	if err != nil {
		return 0, 0, err
	}
	var dummyArr bson.D
	if err = bson.Unmarshal(bts, &dummyArr); err != nil {
		return 0, 0, err
	}
	return len(bts), len(dummyArr), nil
}

func (mh *MonitoringHelper) DBCount(c *mgo.Collection, reqObj interface{}) (int, error) {
	reqSize, err2 := mh.BsonSize(reqObj)
	if err2 != nil {
		return 0, err2
	}
	startTime := time.Now()

	count, err2 := c.Find(reqObj).Count()
	endTime := time.Now()
	mh.submitStoreOpStat2(startTime, endTime, mh.mgoStat, DbOpScan, DbStageRun, err2, reqSize, IntSize, 1)
	if err2 != nil {
		return 0, err2
	}
	return count, nil
}

func (mh *MonitoringHelper) DBScan(c *mgo.Collection, reqObj, result interface{}) error {
	startTime := time.Now()

	err := c.Find(reqObj).All(result)
	endTime := time.Now()

	reqSize, err2 := mh.BsonSize(reqObj)
	if err2 != nil {
		return err2
	}
	respSize, respLen, err2 := mh.BsonSizeLen(result)
	if err2 != nil {
		return err2
	}
	mh.submitStoreOpStat2(startTime, endTime, mh.mgoStat, DbOpScan, DbStageRun, err2, reqSize, respSize, respLen)
	return err
}

func (mh *MonitoringHelper) DBInsert(c *mgo.Collection, reqObj interface{}) error {
	startTime := time.Now()
	err2 := c.Insert(reqObj)
	endTime := time.Now()

	reqSize, err2 := mh.BsonSize(reqObj)
	if err2 != nil {
		return err2
	}
	mh.submitStoreOpStat2(startTime, endTime, mh.mgoStat, DbOpInsert, DbStageRun, err2, reqSize, IntSize, 1)
	if err2 != nil {
		return err2
	}
	return nil
}

func (mh *MonitoringHelper) DBRead(c *mgo.Collection, reqObj, result interface{}) error {
	startTime := time.Now()

	err2 := c.Find(reqObj).One(result)
	endTime := time.Now()

	reqSize, err2 := mh.BsonSize(reqObj)
	if err2 != nil {
		return err2
	}
	respSize, err2 := mh.BsonSize(result)
	if err2 != nil {
		return err2
	}
	mh.submitStoreOpStat2(startTime, endTime, mh.mgoStat, DbOpRead, DbStageRun, err2, reqSize, respSize, 1)
	if err2 != nil {
		return err2
	}
	return nil
}

func (mh *MonitoringHelper) CacheInsert(cli *memcache.Client, it *memcache.Item) error {
	return mh.cacheSet(cli, it, DbOpInsert)
}

func (mh *MonitoringHelper) CacheUpdate(cli *memcache.Client, it *memcache.Item) error {
	return mh.cacheSet(cli, it, DbOpUpdate)
}

func (mh *MonitoringHelper) cacheSet(cli *memcache.Client, it *memcache.Item, op string) error {
	reqSize := len(it.Key) + len(it.Value)
	startTime := time.Now()
	err := cli.Set(it)
	endTime := time.Now()
	mh.submitStoreOpStat2(
		startTime,
		endTime,
		mh.memcStat,
		op,
		DbStageRun,
		err,
		reqSize,
		0,
		1,
	)
	return err
}

func (mh *MonitoringHelper) CacheRead(cli *memcache.Client, key string) (*memcache.Item, error) {
	reqSize := len(key)
	startTime := time.Now()
	resItem, err := cli.Get(key)
	endTime := time.Now()
	respSize := len(resItem.Key) + len(resItem.Value)
	mh.submitStoreOpStat2(
		startTime,
		endTime,
		mh.memcStat,
		DbOpRead,
		DbStageRun,
		err,
		reqSize,
		respSize,
		1,
	)
	return resItem, err
}
