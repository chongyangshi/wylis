package incoming

import (
	"github.com/monzo/typhon"

	"github.com/chongyangshi/wylis/config"
	"github.com/chongyangshi/wylis/metrics"
)

type incomingResponse struct{}

func serveIncoming(req typhon.Request) typhon.Response {
	// After an RPC traffic goes through Calico to the destination pod,
	// the source IP address in the packet will be mangled into the
	// _receiving_ node's Calico IP. Therefore we need to include and
	// retrieve this as a custom header. Authentication of this
	// information is not required in Wylis' use-case.
	sourceNodeIP := req.Header.Get(config.SourceNodeIPHeader)
	metrics.RegisterIncomingRequest(sourceNodeIP)

	// Returns a plain 200 success response.
	return req.Response(incomingResponse{})
}
