// Package types represent shared types
package types

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ConnIDKey connection-id context key, value:string
type ConnIDKey struct{}

// RequestIDKey request-id context key, value:string
type RequestIDKey struct{}

type (
	QueueName       string
	QueueRoutingKey string
	MessageHandler  func(context.Context, amqp.Delivery) error
)
