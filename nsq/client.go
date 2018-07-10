/*
 * Revision History:
 *     Initial: 2018/07/08        Wang RiYu
 */

package nsq

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
)

// Config for nsq.producer
type Config = nsq.Config

// Message for topic
type Message = nsq.Message

// Handler for consumer
type Handler = nsq.Handler

// HandlerFunc for consumer
type HandlerFunc func(message *Message) error

// HandleMessage implements the Handler interface
func (h HandlerFunc) HandleMessage(m *Message) error {
	return h(m)
}

// Client can use a producer for publish and subscribe multi topic/channel message
type Client struct {
	sync.RWMutex
	ID                 int64 // GUID
	Producer           *nsq.Producer
	Topics             map[string]byte          // "topic": client-client'
	Subscribes         map[string]*nsq.Consumer // "topic/channel" -> Consumer
	LookupdHTTPAddress string                   // Lookupd HTTP address
}

// NewClient create an instance of Client
func NewClient(id int64, addr, lookupdHTTPAddress string, config *Config) (*Client, error) {
	if config == nil {
		config = NewConfig()
	}

	producer, err := nsq.NewProducer(addr, config)
	if err != nil {
		return nil, err
	}

	c := &Client{
		ID:                 id,
		LookupdHTTPAddress: lookupdHTTPAddress,
		Producer:           producer,
		Topics:             make(map[string]byte),
		Subscribes:         make(map[string]*nsq.Consumer),
	}

	return c, nil
}

// Publish a message on specified topic
func (c *Client) Publish(topic string, message string) error {
	if topic == "" || message == "" {
		return errors.New("no topic or message")
	}

	if c.Producer == nil {
		return errors.New("no producer")
	}

	err := c.Producer.Publish(topic, []byte(message))
	if err != nil {
		return err
	}

	if _, ok := c.Topics[topic]; !ok {
		c.Topics[topic] = 'a'
	}

	return nil
}

// Subscribe topic/channel message with nsq.consumer
func (c *Client) Subscribe(topic, channel string, config *Config, handler Handler) error {
	var src = strings.Join([]string{topic, channel}, "/")

	if _, ok := c.Subscribes[src]; ok {
		return c.SetHandler(topic, channel, handler)
	}

	if config == nil {
		config = NewConfig()
		config.LookupdPollInterval = 5 * time.Second
	}

	cs, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return err
	}

	cs.SetLogger(nil, 0)

	if handler != nil {
		cs.AddHandler(handler)
	}

	if err := cs.ConnectToNSQLookupd(c.LookupdHTTPAddress); err != nil {
		return err
	}

	c.Subscribes[src] = cs

	return nil
}

// SetHandler set handler of consumer in Client.Subscribes
func (c *Client) SetHandler(topic, channel string, handler Handler) error {
	var src = strings.Join([]string{topic, channel}, "/")

	if _, ok := c.Subscribes[src]; !ok {
		return errors.New(`use Subscribe instead of SetHandler if "topic/channel" is new`)
	}

	if c.Subscribes[src] == nil {
		return errors.New("nil consumer")
	}

	c.Subscribes[src].AddHandler(handler)

	return nil
}

// NewConfig return default Config
func NewConfig() *Config {
	return nsq.NewConfig()
}
