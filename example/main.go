package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/k-yomo/pm"
	"github.com/k-yomo/pm/middleware/logging/pm_zap"
	"github.com/k-yomo/pm/middleware/pm_autoack"
	"github.com/k-yomo/pm/middleware/pm_recovery"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	ctx := context.Background()
	pubsubClient, err := pubsub.NewClient(ctx, "pm-example")
	if err != nil {
		logger.Fatal("initialize pubsub client failed", zap.Error(err))
	}
	defer pubsubClient.Close()

	pubsubPublisher := pm.NewPublisher(
		pubsubClient,
		pm.WithPublishInterceptor(),
	)

	pubsubSubscriber := pm.NewSubscriber(
		pubsubClient,
		pm.WithSubscriptionInterceptor(
			pm_recovery.SubscriptionInterceptor(pm_recovery.WithDebugRecoveryHandler()),
			pm_zap.SubscriptionInterceptor(logger),
			pm_autoack.SubscriptionInterceptor(),
		),
	)
	defer pubsubSubscriber.Close()

	err = pubsubSubscriber.HandleSubscriptionFuncMap(map[string]pm.MessageHandler{
		"example-topic-sub": exampleSubscriptionHandler,
		"example-topic-batch-sub": pm.NewBatchMessageHandler(exampleSubscriptionBatchHandler, pm.BatchMessageHandlerConfig{
			DelayThreshold:    100 * time.Millisecond,
			CountThreshold:    1000,
			ByteThreshold:     1e6,
			BufferedByteLimit: 1e8,
		}),
	})
	if err != nil {
		logger.Fatal("register subscription failed", zap.Error(err))
	}

	pubsubSubscriber.Run(ctx)

	pubsubPublisher.Publish(
		ctx,
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
		return errors.New("error")
	}

	fmt.Println(dataStr)
	return nil
}

func exampleSubscriptionBatchHandler(messages []*pubsub.Message) error {
	batchErr := make(pm.BatchError)
	for _, m := range messages {
		dataStr := string(m.Data)
		if dataStr == "error" {
			batchErr[m.ID] = errors.New("error")
		} else {
			fmt.Println(dataStr)
		}
	}

	return batchErr
}
