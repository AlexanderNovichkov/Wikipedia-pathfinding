package main

import (
	"context"
	"flag"
	"github.com/AlexanderNovichkov/wikipedia-pathfinding/internal/common"
	"github.com/AlexanderNovichkov/wikipedia-pathfinding/internal/pathfinding"
	"github.com/AlexanderNovichkov/wikipedia-pathfinding/pkg/proto/worker"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"
	"log"
	"net/url"
	"time"
)

var (
	requestsQueueUrl     = flag.String("requestsQueueUrl", "amqp://guest:guest@localhost:5672/", "requests queue url")
	resultsStorageDBAddr = flag.String("resultsStorageAddr", "localhost:6379", "results storage address")
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func handleMessage(msg amqp.Delivery, resultStorage *redis.Client) {
	log.Printf("Handle message: %s", msg.Body)

	defer func() {
		if err := recover(); err != nil {
			_ = msg.Nack(false, false)
		}
	}()

	request := worker.FindPathRequestMessage{}
	err := proto.Unmarshal(msg.Body, &request)
	failOnError(err, "Incorrect protobuf message")

	start, err := url.Parse(request.StartPageUrl)
	failOnError(err, "Incorrect start page url: "+request.StartPageUrl)
	finish, err := url.Parse(request.FinishPageUrl)
	failOnError(err, "Incorrect finish page url: "+request.FinishPageUrl)

	path, err := pathfinding.FindPath(*start, *finish)
	failOnError(err, "Failed to find path")

	result := worker.FindPathResultMessage{}
	result.ResultId = request.ResultId
	if len(path) > 0 {
		result.PathFound = true
		for _, currentUrl := range path {
			result.Path = append(result.Path, currentUrl.String())
		}
	}

	log.Println("Saving result:", &result)

	resultMsg, err := proto.Marshal(&result)
	failOnError(err, "Failed to marshal result")

	err = resultStorage.Set(context.Background(), result.ResultId, resultMsg, time.Duration(time.Hour*24)).Err()
	if err != nil {
		log.Println("Failed to save result:", err)
		_ = msg.Nack(false, true)
	} else {
		_ = msg.Ack(false)
	}
}

func main() {
	flag.Parse()

	conn, err := common.ConnectToRabbitMQ(*requestsQueueUrl)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	requestsQueue, err := common.DeclareRequestsQueue(ch)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		requestsQueue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")

	resultsStorage := common.ConnectToResultsStorageDB(*resultsStorageDBAddr)

	for msg := range msgs {
		handleMessage(msg, resultsStorage)
	}

}
