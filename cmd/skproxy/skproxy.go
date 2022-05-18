package main

import (
	"flag"

	"p9t.io/skafos/cmd/skproxy/app"
)

func init() {
	flag.Set("logtostderr", "true")
}

func main() {
	flag.Parse()
	app.StartServer()
}
