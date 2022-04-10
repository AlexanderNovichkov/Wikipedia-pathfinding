package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/AlexanderNovichkov/wikipedia-pathfinding/pkg/proto/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

var (
	addr = flag.String("addr", "localhost:9000", "server address")
)

func readRequest() *server.FindPathRequest {
	request := server.FindPathRequest{}
	fmt.Println("Enter start wikipedia page URL")
	fmt.Scan(&request.StartPageUrl)
	fmt.Println("Enter finish wikipedia page URL")
	fmt.Scan(&request.FinishPageUrl)
	return &request
}

func handleRequest(request *server.FindPathRequest, client server.WikipediaPathfindingClient) {
	fmt.Println("Sending request to server...")
	resultId, err := client.QueueFindPath(context.Background(), request)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ResultId =", resultId.ResultId)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		result, err := client.GetResult(context.Background(), resultId)
		if err == nil {
			printResult(result)
			return
		}

		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			fmt.Println("Request is still being processed")
		} else {
			fmt.Println(err)
			return
		}
	}
}

func printResult(result *server.FindPathResult) {
	if !result.PathFound {
		fmt.Println("Hyperlink path not found")
		return
	}

	fmt.Println("Distance from start page to finish page:", len(result.Path)-1)
	fmt.Println("Path:")
	for _, pageUrl := range result.Path {
		fmt.Println(pageUrl)
	}
}

func main() {
	flag.Parse()

	fmt.Println("Connecting to", *addr)
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := server.NewWikipediaPathfindingClient(conn)

	for {
		request := readRequest()
		handleRequest(request, client)
		fmt.Println()
	}
}
