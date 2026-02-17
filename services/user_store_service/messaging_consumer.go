package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"log/slog"
	"microservices-demo/common/messaging"
	"microservices-demo/common/messaging/userqueue"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type userConsumer struct {
	rabbitmq   *messaging.RabbitMQ
	usersStore *usersStore
}

type usersStore struct {
	users map[string]userqueue.User
	m     sync.RWMutex
}

func newMessageConsumer(rabbitmq *messaging.RabbitMQ) *userConsumer {
	return &userConsumer{rabbitmq: rabbitmq, usersStore: &usersStore{users: make(map[string]userqueue.User)}}
}

func (uc *userConsumer) Listen(ctx context.Context) error {
	slog.Debug("listening user consumer")
	return uc.rabbitmq.ConsumeMessages(ctx, userqueue.Name, func(_ context.Context, msg amqp.Delivery) error {
		user := userqueue.User{}
		if err := gob.NewDecoder(bytes.NewBuffer(msg.Body)).Decode(&user); err != nil {
			return err
		}

		uc.usersStore.m.Lock()
		uc.usersStore.users[user.ID] = user
		uc.usersStore.m.Unlock()
		return nil
	})
}
