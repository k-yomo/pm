# pm

![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)
![Test Workflow](https://github.com/k-yomo/pubsub_cli/workflows/Test/badge.svg)
[![codecov](https://codecov.io/gh/k-yomo/pm/branch/main/graph/badge.svg)](https://codecov.io/gh/k-yomo/pm)
[![Go Report Card](https://goreportcard.com/badge/k-yomo/pm)](https://goreportcard.com/report/k-yomo/pm)

pm is a thin Cloud Pub/Sub client wrapper which lets you manage publishing / subscribing with pluggable middleware.

## Installation

```sh
go get -u github.com/k-yomo/pm
```

## Example

```go
package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"errors"
	"fmt"
	"github.com/k-yomo/pm"
	"github.com/k-yomo/pm/middleware/logging/pm_zap"
	"github.com/k-yomo/pm/middleware/pm_autoack"
	"github.com/k-yomo/pm/middleware/pm_recovery"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
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

	err = pubsubSubscriber.HandleSubscriptionFunc("example-topic-sub", exampleSubscriptionHandler)
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
		fmt.Println("nack will be called to retry")
		return errors.New("error")
	}

	fmt.Println(dataStr)
	return nil
}
```

## Middlewares

### Core Middleware

pm comes equipped with an optional middleware package named `pm_*`.

#### Publisher middleware

| middleware                                                                                                 | description                                                              |
|------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------|
| [Attributes](https://pkg.go.dev/github.com/k-yomo/pm/middleware/pm_attributes#PublishInterceptor)          | Set custom attributes to all outgoing messages when publish              |

#### Subscription middleware

| middleware                                                                                                 | description                                                              |
|------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------|
| [Auto Ack](https://pkg.go.dev/github.com/k-yomo/pm/middleware/pm_autoack#SubscriptionInterceptor)          | Ack automatically depending on if error is returned when subscribe       |
| [Logging - Zap](https://pkg.go.dev/github.com/k-yomo/pm/middleware/logging/pm_zap#SubscriptionInterceptor) | Emit an informative log when subscription processing finish              |
| [Recovery](https://pkg.go.dev/github.com/k-yomo/pm/middleware/pm_recovery#SubscriptionInterceptor)         | Gracefully recover from panics and prints the stack trace when subscribe |

#### Custom middleware
pm middleware is just wrapping publishing / subscribing process which means you can define your custom middleware as well.
- publisher middleware
```go
func MyPublishInterceptor(attrs map[string]string) pm.PublishInterceptor {
    return func (next pm.MessagePublisher) pm.MessagePublisher {
        return func (ctx context.Context, topic *pubsub.Topic, m *pubsub.Message) *pubsub.PublishResult {
        	// do something before publishing
            result := next(ctx, topic, m)
            // do something after publishing
            return result
        }
    }
}
```

- subscription middleware
```go
func MySubscriptionInterceptor() pm.SubscriptionInterceptor {
    return func(_ *pm.SubscriptionInfo, next pm.MessageHandler) pm.MessageHandler {
        return func(ctx context.Context, m *pubsub.Message) error {
            // do something before subscribing
            err := next(ctx, m)
			// do something after subscribing
			return err
        }
    }
}
```
