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

	CfgKeyInfluxBatchSize     = "InfluxBatchSize"
	CfgKeyInfluxFlushInterval = "InfluxFlushInterval"
	CfgKeyInfluxToken         = "InfluxToken"
	CfgKeyInfluxOrg           = "InfluxOrg"
	CfgKeyInfluxBucket        = "InfluxBucket"
	CfgKeyServiceStat         = "InfluxServiceStat"
	CfgKeyMgoStat             = "InfluxMgoStat"
	CfgKeyMemcStat            = "InfluxMemcStat"

	CfgKeySvrDbConn       = "SVR_DBCONN"
	CfgKeySvrTimeout      = "SVR_TIMEOUT"
	CfgKeySvrMemcIdleConn = "SVR_MEMC_IDLE_CONN"
	CfgKeySvrMemcTimeout  = "SVR_MEMC_TIMEOUT"
)
