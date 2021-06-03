package common

const (
	ServiceFrontend  = "frontend"
	ServiceGeo       = "geo"
	ServiceProfile   = "profile"
	ServiceRate      = "rate"
	ServiceRecommend = "recommend"
	ServiceReserve   = "reserve"
	ServiceSearch    = "search"
	ServiceUser      = "user"

	ConfKeyFrontendPort  = "FrontendPort"
	ConfKeyGeoPort       = "GeoPort"
	ConfKeyProfilePort   = "ProfilePort"
	ConfKeyRatePort      = "RatePort"
	ConfKeyRecommendPort = "RecommendPort"
	ConfKeyReservePort   = "ReservePort"
	ConfKeySearchPort    = "SearchPort"
	ConfKeyUserPort      = "UserPort"
)

func GetK8sServiceAddr(name string) {

}
