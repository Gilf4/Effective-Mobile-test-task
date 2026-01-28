package db

import "errors"

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrNoSubscriptionsFound = errors.New("no subscriptions found for the given criteria")
)
