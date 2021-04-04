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

	pubsubManager := pm.NewPubSubManager(
		pubsubClient,
		pm.WithSubscriptionInterceptor(
			pm_recovery.SubscriptionInterceptor,
			pm_autoack.SubscriptionInterceptor,
		),
	)
	defer pubsubManager.Close()

	err = pubsubManager.HandleSubscriptionFunc("pm-example-sub", exampleSubscriptionHandler)
	if err != nil {
		panic(err)
	}

	pubsubManager.Run()

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
		return errors.New("error")
	}

	fmt.Println(string(m.Data))
	return nil
}
