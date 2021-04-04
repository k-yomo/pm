package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"errors"
	"fmt"
	"github.com/k-yomo/pm"
	"github.com/k-yomo/pm/middleware/pm_autoack"
	"github.com/k-yomo/pm/middleware/pm_recovery"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	pubsubClient, err := pubsub.NewClient(context.Background(), "pm-example")
	if err != nil {
		panic(err)
	}
	defer pubsubClient.Close()

	pubsubPublisher := pm.NewPublisher(
		pubsubClient,
		pm.WithPublishInterceptor(
		),
	)

	pubsubSubscriber := pm.NewSubscriber(
		pubsubClient,
		pm.WithSubscriptionInterceptor(
			pm_recovery.SubscriptionInterceptor,
			pm_autoack.SubscriptionInterceptor,
		),
	)
	defer pubsubSubscriber.Close()

	err = pubsubSubscriber.HandleSubscriptionFunc("example-topic-sub", exampleSubscriptionHandler)
	if err != nil {
		panic(err)
	}

	pubsubSubscriber.Run()

	pubsubPublisher.Publish(
		context.Background(),
		pubsubPublisher.Topic("example-topic"),
		&pubsub.Message{
			Data: []byte("test"),
		},
	)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func exampleSubscriptionHandler(ctx context.Context, m *pubsub.Message) error {
	dataStr := string(m.Data)
	if dataStr == "panic" {
		panic("panic")
	}

	if dataStr == "error" {
		fmt.Println("nack will be called to retry")
		return errors.New("error")
	}

	fmt.Println(dataStr)
	return nil
}
