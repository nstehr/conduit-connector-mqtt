package main

import (
	sdk "github.com/conduitio/conduit-connector-sdk"
	mqtt "github.com/nstehr/conduit-connector-mqtt"
)

func main() {
	sdk.Serve(mqtt.Connector)
}
