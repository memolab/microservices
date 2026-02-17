package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"microservices-demo/common/messaging"
	"microservices-demo/common/messaging/userqueue"
	"microservices-demo/common/rid"

	amqp "github.com/rabbitmq/amqp091-go"
)

type messagePublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func newMessagePublisher(rabbitmq *messaging.RabbitMQ) *messagePublisher {
	return &messagePublisher{rabbitmq: rabbitmq}
}

func (up *messagePublisher) publishUser(ctx context.Context, user userqueue.User) error {
	var msg bytes.Buffer
	if err := gob.NewEncoder(&msg).Encode(user); err != nil {
		return err
	}

	return up.rabbitmq.PublishMessage(ctx, userqueue.UserSetData, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/x-gob",
		MessageId:    rid.New16(),
		Body:         msg.Bytes(),
	})
}
