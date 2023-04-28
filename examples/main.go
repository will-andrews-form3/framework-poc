package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/will-andrews-form3/framework-poc"
)

const (
	serverURL = "localhost:4222"
)

func main() {
	natsClient := framework.NewNatsClient(serverURL)

	natsSubscription := framework.NewNatsSubscription(natsClient)

	service := framework.NewService([]framework.Component{natsClient, natsSubscription})

	service.RegisterDependentComponents(natsSubscription, natsClient)

	err := service.Start(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	go sendSomeMessages()

	time.Sleep(time.Second * 10)

	service.Stop(context.Background())
}

func sendSomeMessages() {
	nc, err := nats.Connect(serverURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < 15; i++ {
		_, err := js.Publish("subj", []byte(fmt.Sprintf("message %v", i)))
		if err != nil {
			fmt.Println(err)
			continue
		}
		time.Sleep(time.Second)
	}
}
