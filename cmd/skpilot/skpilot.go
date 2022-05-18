package main

import (
	"flag"

	"p9t.io/skafos/cmd/skpilot/app"
)

func init() {
	flag.Set("logtostderr", "true")
}

func main() {
	flag.Parse()
	app.StartServer()
}
