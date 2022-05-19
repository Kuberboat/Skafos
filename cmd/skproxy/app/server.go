package app

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"p9t.io/skafos/pkg/skproxy"
)

func Listen(port uint16, handler func(resp http.ResponseWriter, req *http.Request)) {
	addr := fmt.Sprintf("0.0.0.0:%v", port)
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	glog.Infof("listening at port %v", port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		glog.Fatal(err)
	}
}

func StartServer() {
	go Listen(skproxy.ProxyPort, skproxy.ProxyRequest)
	go Listen(skproxy.ConfigPort, skproxy.SetConfig)
	select {}
}
