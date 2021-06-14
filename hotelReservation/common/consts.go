package common

const (
	ServiceFrontend = "srv-frontend"
	ServiceGeo      = "srv-geo"
	ServiceProfile  = "srv-profile"
	ServiceRate     = "srv-rate"
	ServiceReco     = "srv-recommendation"
	ServiceResv     = "srv-reservation"
	ServiceSearch   = "srv-search"
	ServiceUser     = "srv-user"

	CfgKeyInfluxBatchSize     = "InfluxBatchSize"
	CfgKeyInfluxFlushInterval = "InfluxFlushInterval"
	CfgKeyInfluxToken         = "InfluxToken"
	CfgKeyInfluxOrg           = "InfluxOrg"
	CfgKeyInfluxBucket        = "InfluxBucket"
	CfgKeyServiceStat         = "InfluxServiceStat"
	CfgKeyMgoStat             = "InfluxMgoStat"
	CfgKeyMemcStat            = "InfluxMemcStat"
)
