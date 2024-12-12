package mqtt

//go:generate paramgen -output=paramgen_src.go SourceConfig

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/conduitio/conduit-commons/config"
	"github.com/conduitio/conduit-commons/opencdc"
	sdk "github.com/conduitio/conduit-connector-sdk"
)

var errQos = errors.New("qosError")

type Source struct {
	sdk.UnimplementedSource

	config SourceConfig
	client *client
}

type SourceConfig struct {
	// Config includes parameters that are the same in the source and destination.
	Config

	Broker   string `json:"broker" validate:"required"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Topic    string `json:"topic" default:"#"`
	Port     int    `json:"port" default:"1883"`
	QOS      int    `json:"qos" default:"0"`
	ClientID string `json:"clientId" default:"mqtt_conduit_client"`
}

type Position struct {
	Topic string `json:"topic"`
}

func (p Position) toSdkPosition() opencdc.Position {
	ps, err := json.Marshal(p)
	// from rabbitmq conenector: https://github.com/conduitio-labs/conduit-connector-rabbitmq/blob/b746573fbbedc4a6b949817a0815ad45b523fd8a/utils.go#L35
	if err != nil {
		// this error should not be possible
		panic(fmt.Errorf("error marshaling position to JSON: %w", err))
	}
	return ps
}

func NewSource() sdk.Source {
	// Create Source and wrap it in the default middleware.
	return sdk.SourceWithMiddleware(&Source{}, sdk.DefaultSourceMiddleware()...)
}

func (s *Source) Parameters() config.Parameters {
	// Parameters is a map of named Parameters that describe how to configure
	// the Source. Parameters can be generated from SourceConfig with paramgen.
	return s.config.Parameters()
}

func (s *Source) Configure(ctx context.Context, cfg config.Config) error {
	// Configure is the first function to be called in a connector. It provides
	// the connector with the configuration that can be validated and stored.
	// In case the configuration is not valid it should return an error.
	// Testing if your connector can reach the configured data source should be
	// done in Open, not in Configure.
	// The SDK will validate the configuration and populate default values
	// before calling Configure. If you need to do more complex validations you
	// can do them manually here.
	err := sdk.Util.ParseConfig(ctx, cfg, &s.config, NewSource().Parameters())
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if s.config.QOS > 2 {
		return fmt.Errorf("%w: %d is not a valid QOS parameter", errQos, s.config.QOS)
	}
	return nil
}

func (s *Source) Open(ctx context.Context, _ opencdc.Position) error {
	// Open is called after Configure to signal the plugin it can prepare to
	// start producing records. If needed, the plugin should open connections in
	// this function. The position parameter will contain the position of the
	// last record that was successfully processed, Source should therefore
	// start producing records after this position. The context passed to Open
	// will be cancelled once the plugin receives a stop signal from Conduit.
	client := newClient(s.config)
	s.client = client
	err := client.connect(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *Source) Read(ctx context.Context) (opencdc.Record, error) {
	// Read returns a new Record and is supposed to block until there is either
	// a new record or the context gets cancelled. It can also return the error
	// ErrBackoffRetry to signal to the SDK it should call Read again with a
	// backoff retry.
	// If Read receives a cancelled context or the context is cancelled while
	// Read is running it must stop retrieving new records from the source
	// system and start returning records that have already been buffered. If
	// there are no buffered records left Read must return the context error to
	// signal a graceful stop. If Read returns ErrBackoffRetry while the context
	// is cancelled it will also signal that there are no records left and Read
	// won't be called again.
	// After Read returns an error the function won't be called again (except if
	// the error is ErrBackoffRetry, as mentioned above).
	// Read can be called concurrently with Ack.
	rec := opencdc.Record{}
	select {
	case <-ctx.Done():
		err := ctx.Err()
		if err != nil {
			return rec, err
		}
		return rec, nil
	case msg := <-s.client.stream():
		var msgKey string
		msgID := msg.MessageID()
		// in my limited testing, all message ids were coming in as 0.
		// attempt here to give them something unique
		if msgID <= 0 {
			msgKey = fmt.Sprintf("%d", time.Now().Unix())
		} else {
			msgKey = strconv.FormatUint(uint64(msgID), 10)
		}
		var (
			pos = Position{
				Topic: msg.Topic(),
			}
			sdkPos   = pos.toSdkPosition()
			metadata = opencdc.Metadata{
				"mqtt.topicName": msg.Topic(),
				"mqtt.qos":       strconv.Itoa(int(msg.Qos())),
			}
			key     = opencdc.RawData([]byte(msgKey))
			payload = opencdc.RawData(msg.Payload())
		)
		rec = sdk.Util.Source.NewRecordCreate(sdkPos, metadata, key, payload)
		return rec, nil
	}
}

func (s *Source) Ack(_ context.Context, _ opencdc.Position) error {
	// Ack signals to the implementation that the record with the supplied
	// position was successfully processed. This method might be called after
	// the context of Read is already cancelled, since there might be
	// outstanding acks that need to be delivered. When Teardown is called it is
	// guaranteed there won't be any more calls to Ack.
	// Ack can be called concurrently with Read.
	return nil
}

func (s *Source) Teardown(_ context.Context) error {
	// Teardown signals to the plugin that there will be no more calls to any
	// other function. After Teardown returns, the plugin should be ready for a
	// graceful shutdown.
	s.client.close()
	return nil
}
