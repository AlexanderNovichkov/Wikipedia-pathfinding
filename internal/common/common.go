package common

import (
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

func DeclareRequestsQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"pathfinding_requests",
		false,
		false,
		false,
		false,
		nil,
	)
}

func ConnectToResultsStorageDB(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}
