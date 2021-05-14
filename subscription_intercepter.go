package pm

// SubscriptionInfo contains various info about the subscription.
type SubscriptionInfo struct {
	TopicID        string
	SubscriptionID string
}

// SubscriptionInterceptor provides a hook to intercept the execution of a message handling.
type SubscriptionInterceptor = func(info *SubscriptionInfo, next MessageHandler) MessageHandler
