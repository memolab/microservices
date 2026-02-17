package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"microservices-demo/common/messaging/userqueue"
	"microservices-demo/common/retry"
	"microservices-demo/common/types"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	// AppExchange app exchange name
	AppExchange = "appx"
	// DeadLetterExchange dead-latter exchange name
	DeadLetterExchange = "dlx"
	// DeadLetterQueue dead-latter queue name
	DeadLetterQueue = "dlq"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
}

// NewRabbitMQ return new rabbitmq instance and setup queues by:
//
// bind queues to its routing keys in app exchange
//
// also bind dead letter queue in it`s exchange
func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create channel: %v", err)
	}

	rmq := &RabbitMQ{
		conn:    conn,
		Channel: ch,
	}

	if err := rmq.setupExchangesAndQueues(); err != nil {
		rmq.Close()
		return nil, fmt.Errorf("failed to setup exchanges and queues: %v", err)
	}

	return rmq, nil
}

// Close close channel and conn
func (r *RabbitMQ) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

// ConsumeMessages in AppExchange it`s recive queue channel messages and pass every to handler for process with handler(newctx context, msg amqp.Delivery)
// useing signal context to stop listening
func (r *RabbitMQ) ConsumeMessages(sctx context.Context, queueName types.QueueName, handler types.MessageHandler) error {
	// Set prefetch count to 1 for fair dispatch
	// This tells RabbitMQ not to give more than one message to a service at a time.
	// The worker will only get the next message after it has acknowledged the previous one.
	err := r.Channel.Qos(
		1,     // prefetchCount: Limit to 1 unacknowledged message per consumer
		0,     // prefetchSize: No specific limit on message size
		false, // global: Apply prefetchCount to each consumer individually
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %v", err)
	}

	msgs, err := r.Channel.Consume(
		string(queueName), // queue
		"",                // consumer
		false,             // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		return err
	}

	go func() {
		for message := range msgs {
			slog.Debug("message received", "routing_key", message.RoutingKey, "message_id", message.MessageId)
			// Extract trace context from message headers
			carrier := amqpHeadersCarrier(message.Headers)
			trctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)
			// trace Consume operation
			tracer := otel.GetTracerProvider().Tracer("rabbitmq")
			trctx, span := tracer.Start(trctx, "rabbitmq.consume",
				trace.WithAttributes(
					attribute.String("messaging.destination", message.Exchange),
					attribute.String("messaging.routing_key", message.RoutingKey),
					attribute.String("messaging.message_id", message.MessageId),
				),
			)

			if err := handler(trctx, message); err != nil {
				slog.Error("fail processing message", "error", err, "routing_key", message.RoutingKey, "message_id", message.MessageId)
				span.SetStatus(codes.Error, err.Error())
				span.End()
				// Add failure context before sending to the DLQ
				headers := amqp.Table{}
				if message.Headers != nil {
					headers = message.Headers
				}
				headers["x-death-reason"] = err.Error()
				headers["x-origin-exchange"] = message.Exchange
				headers["x-original-routing-key"] = message.RoutingKey
				message.Headers = headers
				// Reject without requeue - message will go to the DLQ
				_ = message.Reject(false)
				continue
			} else {
				span.SetStatus(codes.Ok, "message processed")
				span.End()
			}

			// Only Ack if the handler succeeds
			if ackErr := message.Ack(false); ackErr != nil {
				slog.Error("fail to ack message", "error", ackErr, "routing_key", message.RoutingKey, "message_id", message.MessageId)
			}
			slog.Debug("message processed", "routing_key", message.RoutingKey, "message_id", message.MessageId)

			select {
			case <-sctx.Done():
				return
			default:
			}
		}
	}()

	return nil
}

// PublishMessage in AppExchange try publishing on routing key
func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey types.QueueRoutingKey, message amqp.Publishing) error {
	// trace publish operation
	tracer := otel.GetTracerProvider().Tracer("rabbitmq")
	ctx, span := tracer.Start(ctx, "rabbitmq.publish",
		trace.WithAttributes(
			attribute.String("messaging.destination", AppExchange),
			attribute.String("messaging.routing_key", string(routingKey)),
			attribute.String("messaging.message_id", message.MessageId),
		),
	)
	defer span.End()

	// Inject trace context into message headers
	if message.Headers == nil {
		message.Headers = make(amqp.Table)
	}
	carrier := amqpHeadersCarrier(message.Headers)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	message.Headers = amqp.Table(carrier)

	// publish with retry
	cfg := retry.Config{MaxRetries: 3, InitialWait: 300 * time.Millisecond, MaxWait: 1500 * time.Millisecond}
	err := retry.WithBackoff(ctx, cfg, func() error {
		return r.Channel.PublishWithContext(ctx,
			AppExchange,        // exchange
			string(routingKey), // routing key
			false,              // mandatory
			false,              // immediate
			message,
		)
	})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.Error("fail publishing mesage", "error", err, "retries", cfg.MaxRetries, "routing_key", routingKey, "message_id", message.MessageId)
		return err
	}
	span.SetStatus(codes.Ok, "message published")
	slog.Debug("message published", "routing_key", routingKey, "message_id", message.MessageId)
	return nil
}

func (r *RabbitMQ) setupDeadLetterExchange() error {
	// Declare the dead letter exchange
	if err := r.Channel.ExchangeDeclare(
		DeadLetterExchange,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	); err != nil {
		return err
	}

	// Declare the dead letter queue
	q, err := r.Channel.QueueDeclare(
		DeadLetterQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	// Bind the queue to the exchange with a wildcard routing key
	if err = r.Channel.QueueBind(
		q.Name,
		"#", // wildcard routing key to catch all messages
		DeadLetterExchange,
		false,
		nil,
	); err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) setupExchangesAndQueues() error {
	// First setup the DLQ exchange and queue
	if err := r.setupDeadLetterExchange(); err != nil {
		return err
	}

	if err := r.Channel.ExchangeDeclare(
		AppExchange, // name
		"topic",     // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(userqueue.Name, userqueue.GetRoutingKeys()); err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) declareAndBindQueue(queueName types.QueueName, messageKeys []types.QueueRoutingKey) error {
	// Add dead letter configuration
	args := amqp.Table{
		"x-dead-letter-exchange": DeadLetterExchange,
	}

	q, err := r.Channel.QueueDeclare(
		string(queueName), // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		args,              // arguments with DLX config
	)
	if err != nil {
		return err
	}

	for _, rk := range messageKeys {
		if err := r.Channel.QueueBind(
			q.Name,      // queue name
			string(rk),  // routing key
			AppExchange, // exchange
			false,
			nil,
		); err != nil {
			return err
		}
	}

	return nil
}

// amqpHeadersCarrier implements the TextMapCarrier interface for AMQP headers
type amqpHeadersCarrier amqp.Table

func (c amqpHeadersCarrier) Get(key string) string {
	if v, ok := c[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c amqpHeadersCarrier) Set(key string, value string) {
	c[key] = value
}

func (c amqpHeadersCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}
