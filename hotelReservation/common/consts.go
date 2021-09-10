package common

const (
	ServiceFrontend = "frontend"
	ServiceGeo      = "geo"
	ServiceProfile  = "profile"
	ServiceRate     = "rate"
	ServiceReco     = "recommendation"
	ServiceResv     = "reservation"
	ServiceSearch   = "search"
	ServiceUser     = "user"

	ServiceMemcProfile = "memcached-profile"
	ServiceMemcRate    = "memcached-rate"
	ServiceMemcResv    = "memcached-reserve"

	CfgKeyInfluxBatchSize     = "InfluxBatchSize"
	CfgKeyInfluxFlushInterval = "InfluxFlushInterval"
	CfgKeyInfluxToken         = "InfluxToken"
	CfgKeyInfluxOrg           = "InfluxOrg"
	CfgKeyInfluxBucket        = "InfluxBucket"

	CfgKeyServiceStat = "InfluxServiceStat"
	CfgKeyMgoStat     = "InfluxMgoStat"
	CfgKeyMemcStat    = "InfluxMemcStat"

	CfgKeySvrDbConn       = "SVR_DBCONN"
	CfgKeySvrTimeout      = "SVR_TIMEOUT"
	CfgKeySvrMemcIdleConn = "SVR_MEMC_IDLE_CONN"
	CfgKeySvrMemcTimeout  = "SVR_MEMC_TIMEOUT"

	MongoPort     = 27017
	MemcachedPort = 11211

	EnvMyPodName = "MY_POD_NAME"
	EnvMyPodIp   = "MY_POD_IP"
	EnvMyPodNs   = "MY_POD_NS"

	EnvParamCenterAddr = "ParamCenterAddr"
	EnvParamCenterPort = "ParamCenterPort"

	PCenterPathAcquireParam = "/params"

	ParamTypeKernel = "k"
	ParamTypeConfig = "c"

	ParamEnvTplKernel = "KERNEL_%s_%d"
	ParamEnvTplConfig = "CONFIG_%s_%d"
)
