package main

import (
	"context"
	"log"

	"github.com/na0chan-go/grpc-lesson/pb"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewFileServiceClient(conn)
	ctx := context.Background()
	callListFiles(c, ctx)
}

func callListFiles(c pb.FileServiceClient, ctx context.Context) {
	res, err := c.ListFiles(ctx, &pb.ListFilesRequest{})
	if err != nil {
		log.Fatalf("Failed to call ListFiles: %v", err)
	}
	log.Printf("Response from ListFiles: %v", res.Filenames)
}
