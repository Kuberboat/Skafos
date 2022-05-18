package app

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"p9t.io/skafos/pkg/skproxy"
)

func ListenProxy() {
	addr := fmt.Sprintf("0.0.0.0:%v", skproxy.ProxyPort)
	mux := http.NewServeMux()
	mux.HandleFunc("/", skproxy.ProxyRequest)
	glog.Infof("listening at port %v", skproxy.ProxyPort)
	if err := http.ListenAndServe(addr, mux); err != nil {
		glog.Fatal(err)
	}
}

func StartServer() {
	go ListenProxy()
	select {}
}
