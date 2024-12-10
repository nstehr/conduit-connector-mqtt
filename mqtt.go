package mqtt

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type client struct {
	mqtt   mqtt.Client
	config SourceConfig
	queue  *queue
}

func newClient(config SourceConfig) *client {
	opts := mqtt.NewClientOptions()

	opts.AddBroker(fmt.Sprintf("ssl://%s", net.JoinHostPort(config.Broker, strconv.Itoa(config.Port))))
	opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: true}) // #nosec G402
	opts.SetKeepAlive(60 * time.Second)
	opts.SetConnectTimeout(60 * time.Second)
	opts.SetClientID(config.ClientID)
	opts.SetUsername(config.Username)
	opts.SetPassword(config.Password)
	mqtt := mqtt.NewClient(opts)
	return &client{mqtt: mqtt, config: config, queue: newQueue()}
}

func (c *client) connect(ctx context.Context) error {
	sdk.Logger(ctx).Debug().Msg("mqtt client connecting...")
	token := c.mqtt.Connect()
	// this will wait indefinitely. There is a timeout version
	// but it seemed to exit instantly without actually waiting the timeout
	token.Wait()

	if token.Error() != nil {
		return fmt.Errorf("mqtt broker connector error: %w", token.Error())
	}

	filters := make(map[string]byte)
	for _, topic := range c.config.Topics {
		filters[topic] = byte(c.config.QOS)
	}

	token = c.mqtt.SubscribeMultiple(filters, func(_ mqtt.Client, msg mqtt.Message) {
		c.queue.push(msg)
	})
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("mqtt subscribe error: %w", token.Error())
	}
	return nil
}

func (c *client) read(ctx context.Context) (mqtt.Message, bool) {
	sdk.Logger(ctx).Debug().Msg("Reading next mqtt message")
	return c.queue.next()
}

type queue struct {
	mu   sync.Mutex
	data []mqtt.Message
	cond *sync.Cond
}

func newQueue() *queue {
	q := &queue{data: make([]mqtt.Message, 0)}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *queue) push(item mqtt.Message) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.data = append(q.data, item)
	q.cond.Signal()
}

func (q *queue) next() (mqtt.Message, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.data) == 0 {
		q.cond.Wait()
	}
	item := q.data[0]
	q.data = q.data[1:]
	return item, true
}
