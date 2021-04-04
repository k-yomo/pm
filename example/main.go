package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/k-yomo/pm"
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
		pm.WithSubscriptionInterceptor(pm_recovery.SubscriptionInterceptor),
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

func exampleSubscriptionHandler(ctx context.Context, m *pubsub.Message)  {
	if string(m.Data) == "panic" {
		panic("panic")
	} else {
		fmt.Println(string(m.Data))
	}
}