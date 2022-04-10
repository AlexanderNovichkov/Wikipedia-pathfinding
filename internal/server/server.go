package server

import (
	"context"
	"github.com/AlexanderNovichkov/wikipedia-pathfinding/pkg/proto/server"
	"github.com/AlexanderNovichkov/wikipedia-pathfinding/pkg/proto/worker"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"log"
)

type Server struct {
	server.UnimplementedWikipediaPathfindingServer
	ch             *amqp.Channel
	requestsQueue  *amqp.Queue
	resultsStorage *redis.Client
}

func NewServer(ch *amqp.Channel, q *amqp.Queue, resultsStorage *redis.Client) *Server {
	return &Server{ch: ch, requestsQueue: q, resultsStorage: resultsStorage}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func (s *Server) QueueFindPath(ctx context.Context, request *server.FindPathRequest) (*server.FindPathResultId, error) {
	workerRequestMessage := worker.FindPathRequestMessage{
		StartPageUrl:  request.StartPageUrl,
		FinishPageUrl: request.FinishPageUrl,
		ResultId:      uuid.New().String(),
	}
	workerRequestMessageEncoded, err := proto.Marshal(&workerRequestMessage)
	failOnError(err, "Failed to encode worker request message")

	err = s.ch.Publish(
		"",
		s.requestsQueue.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         workerRequestMessageEncoded,
		})
	failOnError(err, "Failed to publish message")

	return &server.FindPathResultId{
		ResultId: workerRequestMessage.ResultId,
	}, nil
}

func (s *Server) GetResult(ctx context.Context, request *server.FindPathResultId) (*server.FindPathResult, error) {
	resultMsg, err := s.resultsStorage.Get(ctx, request.ResultId).Bytes()
	if err == redis.Nil {
		return nil, status.Error(codes.NotFound, "Request is still being processed or FindPathResultId is incorrect")
	}
	if err != nil {
		log.Println("Failed to get result from resultsStorage:", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	var result worker.FindPathResultMessage
	if proto.Unmarshal(resultMsg, &result) != nil {
		log.Println("Failed to unmarshal result message:", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &server.FindPathResult{
		PathFound: result.PathFound,
		Path:      result.Path,
	}, nil
}
