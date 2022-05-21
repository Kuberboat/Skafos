package main

import (
	"flag"

	"p9t.io/skafos/cmd/skagent/app"
	"p9t.io/skafos/pkg/api/core"
)

var (
	address        string
	port           uint
	skPilotAddress string
	skPilotPort    uint
)

func init() {
	flag.Set("logtostderr", "true")
	flag.StringVar(&address, "host-ip", "localhost", "IPv4 address of this agent skpilot will see when this agent is registered.")
	flag.UintVar(&port, "port", core.SKAGENT_PORT, "Port skagent listens to.")
	flag.StringVar(&skPilotAddress, "skpilot-ip", "localhost", "IPv4 address of the host skpilot runs on.")
	flag.UintVar(&skPilotPort, "skpilot-port", core.SKPILOT_PORT, "Port skpilot listens to.")
}

func main() {
	flag.Parse()
	app.StartServer(address, uint16(port), skPilotAddress, uint16(skPilotPort))
}
