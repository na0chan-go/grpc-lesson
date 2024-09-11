package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/na0chan-go/grpc-lesson/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedFileServiceServer // これを埋め込むことで、未実装のメソッドを持つサーバーを作成することができる
}

func (s *server) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	fmt.Println("ListFiles was invoked")
	dir := "./storage"
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	filenames := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			filenames = append(filenames, file.Name())
		}
	}

	res := &pb.ListFilesResponse{
		Filenames: filenames,
	}

	return res, nil
}

func (s *server) Download(req *pb.DownloadRequest, stream pb.FileService_DownloadServer) error {
	fmt.Println("Download was invoked")

	filename := req.GetFilename()
	path := "./storage/" + filename

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return status.Errorf(codes.NotFound, "file not found: %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, 5)
	for {
		n, err := file.Read(buf)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		res := &pb.DownloadResponse{
			Data: buf[:n],
		}
		if err := stream.Send(res); err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (s *server) Upload(stream pb.FileService_UploadServer) error {
	fmt.Println("Upload was invoked")

	var buf bytes.Buffer
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			res := &pb.UploadResponse{
				Size: int32(buf.Len()),
			}
			return stream.SendAndClose(res)
		}
		if err != nil {
			return err
		}
		data := req.GetData()
		log.Printf("Received data(bytes): %v", data)
		log.Printf("Received data(string): %v", string(data))
		buf.Write(data)
	}
}

func (s *server) UploadAndNotifyProgress(stream pb.FileService_UploadAndNotifyProgressServer) error {
	fmt.Println("UploadAndNotifyProgress was invoked")

	size := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		data := req.GetData()
		log.Printf("Received data: %v", data)
		size += len(data)

		res := &pb.UploadAndNotifyProgressResponse{
			Message: fmt.Sprintf("Received %d bytes", size),
		}
		if err := stream.Send(res); err != nil {
			return err
		}
	}

	return nil
}

func myLogging() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		log.Printf("request data: %+v", req)

		resp, err = handler(ctx, req)
		if err != nil {
			return nil, err
		}
		log.Printf("response data: %+v", resp)

		return resp, nil
	}
}

func authorize(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}
	if token != "test-token" {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return ctx, nil
}

func main() {
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(
			myLogging(),
			grpc_auth.UnaryServerInterceptor(authorize),
		)))
	pb.RegisterFileServiceServer(s, &server{})

	fmt.Println("Server is running on port: 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
