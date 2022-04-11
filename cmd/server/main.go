package main

import (
	"flag"
	"github.com/AlexanderNovichkov/wikipedia-pathfinding/internal/common"
	"github.com/AlexanderNovichkov/wikipedia-pathfinding/internal/server"
	serverpb "github.com/AlexanderNovichkov/wikipedia-pathfinding/pkg/proto/server"
	"google.golang.org/grpc"
	"log"
	"net"
)

var (
	requestsQueueUrl     = flag.String("requestsQueueUrl", "amqp://guest:guest@localhost:5672/", "requests queue url")
	resultsStorageDBAddr = flag.String("resultsStorageAddr", "localhost:6379", "results storage address")
	serverAddr           = flag.String("serverAddr", ":9000", "server address")
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
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

	lis, err := net.Listen("tcp", *serverAddr)
	failOnError(err, "Failed to listen")

	resultsStorage := common.ConnectToResultsStorageDB(*resultsStorageDBAddr)

	srv := grpc.NewServer()
	serverpb.RegisterWikipediaPathfindingServer(srv, server.NewServer(ch, &requestsQueue, resultsStorage))

	log.Fatalln(srv.Serve(lis))
}
