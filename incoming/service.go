package incoming

import "github.com/monzo/typhon"

func Service() typhon.Service {
	router := typhon.Router{}
	router.GET("/incoming", serveIncoming)
	router.GET("/healthz", serveHealthCheck)

	// We do not need a client error filter or CORS filter, as this service serves
	// internal traffic from other Wylis pods only.
	svc := router.Serve().Filter(typhon.ErrorFilter).Filter(typhon.H2cFilter)

	return svc
}
