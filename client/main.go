package main

import (
	"context"
	"io"
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
	// callListFiles(c, ctx)
	callDownload(c, ctx)
}

func callListFiles(c pb.FileServiceClient, ctx context.Context) {
	res, err := c.ListFiles(ctx, &pb.ListFilesRequest{})
	if err != nil {
		log.Fatalf("Failed to call ListFiles: %v", err)
	}
	log.Printf("Response from ListFiles: %v", res.Filenames)
}

func callDownload(c pb.FileServiceClient, ctx context.Context) {
	req := &pb.DownloadRequest{
		Filename: "name.txt",
	}
	stream, err := c.Download(ctx, req)
	if err != nil {
		log.Fatalf("Failed to call Download: %v", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive data: %v", err)
		}
		log.Printf("Response from Download(bytes): %v", res.GetData())
		log.Printf("Response from Download(string): %v", string(res.GetData()))
	}
}
