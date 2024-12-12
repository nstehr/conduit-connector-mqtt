package mqtt

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
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
	// mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	// mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	// mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)

	opts := mqtt.NewClientOptions()
	q := newQueue()
	opts.AddBroker(fmt.Sprintf("ssl://%s", net.JoinHostPort(config.Broker, strconv.Itoa(config.Port))))
	opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: true}) // #nosec G402
	opts.SetKeepAlive(60 * time.Second)
	opts.SetConnectTimeout(60 * time.Second)
	opts.SetClientID(config.ClientID)
	opts.SetUsername(config.Username)
	opts.SetPassword(config.Password)

	// setting the subscription up automatically on connect. I've seen cases where the client gets disconnected
	// from the broker, reconnects and then it appears the subscription stops working. This _should_ create a
	// new subscription on reconnect
	opts.OnConnect = func(client mqtt.Client) {
		token := client.Subscribe(config.Topic, byte(config.QOS), func(_ mqtt.Client, msg mqtt.Message) {
			// using a queue here instead of channel to prevent blocking on waiting for a read from the channel in this handler
			// from the docs: 'callback must not block or call functions within this package that may block'
			// the queue does lock to write/read but that should be fairly fast
			q.push(msg)
		})
		token.Wait()
		if token.Error() != nil {
			log.Println("ERROR SUBSCRIBING")
		}
	}

	return &client{mqtt: mqtt.NewClient(opts), config: config, queue: q}
}

func (c *client) connect(ctx context.Context) error {
	sdk.Logger(ctx).Debug().Msg("mqtt client connecting...")
	c.queue.open()
	token := c.mqtt.Connect()
	// this will wait indefinitely. There is a timeout version
	// but it seemed to exit instantly without actually waiting the timeout
	token.Wait()

	if token.Error() != nil {
		return fmt.Errorf("mqtt broker connector error: %w", token.Error())
	}

	return nil
}

func (c *client) stream() <-chan mqtt.Message {
	return c.queue.out
}

func (c *client) close() {
	if c != nil && c.queue != nil {
		c.queue.close()
	}
}

type queue struct {
	mu     sync.Mutex
	data   []mqtt.Message
	cond   *sync.Cond
	closed bool
	out    chan mqtt.Message
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

func (q *queue) close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.cond.Broadcast()
}

func (q *queue) open() {
	q.out = make(chan mqtt.Message)
	go func() {
		defer close(q.out)
		for {
			q.mu.Lock()
			for !q.closed && len(q.data) == 0 {
				q.cond.Wait()
			}
			if q.closed && len(q.data) == 0 {
				q.mu.Unlock()
				return
			}
			item := q.data[0]
			q.data = q.data[1:]
			q.mu.Unlock()

			q.out <- item
		}
	}()
}
