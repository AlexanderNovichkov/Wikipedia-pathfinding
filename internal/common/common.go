package common

import (
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
	"log"
	"time"
)

func ConnectToRabbitMQ(url string) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error
	for {
		log.Println("Connecting to RabbitMq using url = ", url)
		if conn, err = amqp.Dial(url); err == nil {
			break
		}
		log.Println(err)
		time.Sleep(time.Second)
	}
	log.Println("Connected to RabbitMq")
	return conn, nil
}

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
