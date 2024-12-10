package mqtt_test

import (
	"context"
	"testing"

	"github.com/matryer/is"
	mqtt "github.com/nstehr/conduit-connector-mqtt"
)

func TestTeardown_NoOpen(t *testing.T) {
	is := is.New(t)
	con := mqtt.NewDestination()
	err := con.Teardown(context.Background())
	is.NoErr(err)
}
