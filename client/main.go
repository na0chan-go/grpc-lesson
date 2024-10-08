package main

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/na0chan-go/grpc-lesson/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func main() {
	// 証明書を読み込む
	certFile := "/Users/naocha/Library/Application Support/mkcert/rootCA.pem"
	creds, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		log.Fatalf("Failed to load credentials: %v", err)
	}
	// サーバーに接続
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewFileServiceClient(conn)
	ctx := context.Background()
	// callListFiles(c, ctx)
	callDownload(c, ctx)
	// callUpload(c, ctx)
	// callUploadAndNotifyProgress(c, ctx)
}

func callListFiles(c pb.FileServiceClient, ctx context.Context) {
	// ヘッダーを追加
	md := metadata.New(map[string]string{"authorization": "Bearer bad-token"})
	// ヘッダーをコンテキストに追加
	ctx = metadata.NewOutgoingContext(ctx, md)
	res, err := c.ListFiles(ctx, &pb.ListFilesRequest{})
	if err != nil {
		log.Fatalf("Failed to call ListFiles: %v", err)
	}
	log.Printf("Response from ListFiles: %v", res.Filenames)
}

func callDownload(c pb.FileServiceClient, ctx context.Context) {
	// タイムアウトを設定
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

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
			resErr, ok := status.FromError(err)
			if ok {
				if resErr.Code() == codes.NotFound {
					log.Fatalf("Error Code: %v, Error Message: %v", resErr.Code(), resErr.Message())
				} else if resErr.Code() == codes.DeadlineExceeded {
					log.Fatalf("Deadline exceeded: %v", resErr.Message())
				} else {
					log.Fatalf("Unknown grpc error: %v", err)
				}
			} else {
				log.Fatalf("Failed to receive data: %v", err)
			}
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

func callUploadAndNotifyProgress(c pb.FileServiceClient, ctx context.Context) {
	filename := "sports.txt"
	path := "./storage/" + filename
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()
	stream, err := c.UploadAndNotifyProgress(ctx)
	if err != nil {
		log.Fatalf("Failed to call UploadAndNotifyProgress: %v", err)
	}

	// request
	buf := make([]byte, 5)
	// ファイルの読み込みを行う
	go func() {
		for {
			n, err := file.Read(buf)
			if n == 0 || err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Failed to read file: %v", err)
			}
			req := &pb.UploadAndNotifyProgressRequest{
				Data: buf[:n],
			}
			if err := stream.Send(req); err != nil {
				log.Fatalf("Failed to send request: %v", err)
			}
			time.Sleep(1 * time.Second)
		}
		// ファイルの読み込みが終了したらストリームを閉じる
		if err := stream.CloseSend(); err != nil {
			log.Fatalf("Failed to close stream: %v", err)
		}
	}()

	// response
	ch := make(chan struct{})
	// ストリームからのレスポンスを受け取る
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Failed to receive response: %v", err)
			}
			log.Printf("Received message: %v", res.GetMessage())
		}
		// ストリームが閉じられたらチャネルを閉じる
		close(ch)
	}()
	// チャネルが閉じられるまで待つ
	<-ch
}
