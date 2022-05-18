package skproxy

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

const (
	// ProxyPort is the port number on which skproxy receives http requests and forward them.
	ProxyPort uint16 = 16000
)

// getPort extracts the target port from the request. Default to 80.
func getPort(req *http.Request) (uint16, error) {
	portStr := req.URL.Port()
	if len(portStr) == 0 {
		portStr = "80"
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(port), nil
}

// getNewHost looks up the proxy rules and determine which IP/domain and port
// the original request should be forwarded to.
// The returned host will contain the new host and port. Port 80 will not be omitted.
func getNewHost(host string, port uint16) string {
	// Currently just forward as is.
	return fmt.Sprintf("%v:%v", host, port)
}

// buildNewRequest builds a new request based on the original request, except that
// the host and port might be altered based on proxy rules.
func buildNewRequest(req *http.Request) (*http.Request, error) {
	// NOTE: Some fields in newReq are shallow copied, but it seems fine.
	newReq := new(http.Request)
	*newReq = *req

	// Get original port.
	port, err := getPort(req)
	if err != nil {
		return nil, fmt.Errorf("invalid request port: %v", err.Error())
	}

	// Strip off port from original host.
	host := req.Host
	colonIdx := strings.LastIndex(host, ":")
	if colonIdx != -1 {
		host = host[0:colonIdx]
	}

	// Modify new request data.
	newHost := getNewHost(host, port)
	newReq.Host = newHost
	newReq.URL.Host = newHost
	newReq.RequestURI = newReq.URL.String()

	return newReq, nil
}

func ProxyRequest(resp http.ResponseWriter, req *http.Request) {
	transport := http.DefaultTransport

	// Serve http request only.
	if !strings.HasPrefix(req.Proto, "HTTP") {
		resp.WriteHeader(http.StatusBadGateway)
		resp.Write([]byte("only http is supported"))
		return
	}
	req.URL.Scheme = "http"

	// Build new request.
	newReq, err := buildNewRequest(req)
	if err != nil {
		resp.WriteHeader(http.StatusBadGateway)
		resp.Write([]byte(fmt.Sprintf("invalid request port: %v", err.Error())))
		return
	}

	// Send the new request request.
	glog.Infof(fmt.Sprintf("%v -> %v", req.RequestURI, newReq.RequestURI))
	forwardedResp, err := transport.RoundTrip(newReq)
	if err != nil {
		resp.WriteHeader(http.StatusBadGateway)
		return
	}

	// Copy response.
	for k, vs := range forwardedResp.Header {
		for _, v := range vs {
			resp.Header().Add(k, v)
		}
	}
	resp.WriteHeader(forwardedResp.StatusCode)
	io.Copy(resp, forwardedResp.Body)
	forwardedResp.Body.Close()
}