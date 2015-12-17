package dolb

import "net/http"

// LBCreateHandler is a http handler for creating a load balancer.
func LBCreateHandler(r *http.Request) Response {
	output := struct {
		ID string
	}{
		ID: "lb-1",
	}
	return Response{body: output, status: http.StatusCreated}
}
