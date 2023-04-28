package framework

import (
	"context"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

const (
	subject    = "subj"
	streamName = "Mystream"
)

type NatsSubscription struct {
	natsClient *NatsClient
}

func NewNatsSubscription(client *NatsClient) *NatsSubscription {
	return &NatsSubscription{
		natsClient: client,
	}
}

func (n *NatsSubscription) Start(ctx context.Context) error {
	fmt.Println("nats sub starting")

	err := createStream(n.natsClient.jsContext)
	if err != nil {
		return err
	}

	go subscriber(n.natsClient.jsContext)

	return nil
}

func (n *NatsSubscription) Stop(ctx context.Context) error {
	fmt.Println("nats sub stopping")
	return nil
}

func createStream(js nats.JetStreamContext) error {
	stream, err := js.StreamInfo(streamName)
	if err != nil && !strings.Contains(err.Error(), "stream not found") {
		return errors.Wrap(err, "failed to check if stream existed")
	}

	// stream exists already
	if stream != nil {
		fmt.Println("stream already exists")
		return nil
	}

	fmt.Println("stream doesn't exist so creating")

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "test",
		Subjects: []string{subject},
	})

	if err != nil {
		return errors.Wrap(err, "failed to create stream")
	}
	return nil
}

func subscriber(js nats.JetStreamContext) {
	_, err := js.Subscribe(subject, func(msg *nats.Msg) {
		err := msg.Ack()
		if err != nil {
			fmt.Printf("failed to ack message: %s\n", err)
			return
		}
		fmt.Printf("message received: %s\n", msg.Data)
	}, nats.DeliverNew())

	if err != nil {
		fmt.Printf("failed to subscribe: %s\n", err)
	}
}
