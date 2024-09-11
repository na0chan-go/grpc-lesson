package main

import (
	"context"
	"io"
	"log"
	"os"
	"time"

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
	// callDownload(c, ctx)
	callUpload(c, ctx)
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

func callUpload(c pb.FileServiceClient, ctx context.Context) {
	filename := "sports.txt"
	path := "./storage/" + filename
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()
	stream, err := c.Upload(ctx)
	if err != nil {
		log.Fatalf("Failed to call Upload: %v", err)
	}

	buf := make([]byte, 5)
	for {
		n, err := file.Read(buf)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		req := &pb.UploadRequest{
			Data: buf[:n],
		}
		if err := stream.Send(req); err != nil {
			log.Fatalf("Failed to send request: %v", err)
		}
		time.Sleep(1 * time.Second)
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Failed to receive response: %v", err)
	}
	log.Printf("Received data size: %v", res.GetSize())
}
