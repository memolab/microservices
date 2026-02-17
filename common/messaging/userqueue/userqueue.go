// Package userqueue all related to the user queue, including routing keys and messages structs to be shared with producer and consumer
package userqueue

import "microservices-demo/common/types"

// const of globaly unique queue name and queue routing keys
const (
	Name        types.QueueName       = "users"
	UserSetData types.QueueRoutingKey = "users_set_data"
)

type User struct {
	ID   string
	Name string
	Data string
}

func GetRoutingKeys() []types.QueueRoutingKey {
	return []types.QueueRoutingKey{
		UserSetData,
	}
}
