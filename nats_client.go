package framework

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

type NatsClient struct {
	serverURL string
	client    *nats.Conn
	jsContext nats.JetStreamContext
}

func NewNatsClient(serverURL string) *NatsClient {
	return &NatsClient{
		serverURL: serverURL,
	}
}

func (n *NatsClient) Start(ctx context.Context) error {
	fmt.Println("nats client starting")

	nc, err := nats.Connect(n.serverURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to NATs")
	}

	n.client = nc

	js, err := nc.JetStream()
	if err != nil {
		return errors.Wrap(err, "failed to setup Jetstream")
	}
	n.jsContext = js

	return nil
}

func (n *NatsClient) Stop(ctx context.Context) error {
	fmt.Println("nats client stopping")

	n.client.Close()

	return nil
}
