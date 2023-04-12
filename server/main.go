package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	pb "app/docpdf"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedConvertServiceServer
}

func (s *server) ConvertToPdf(ctx context.Context, req *pb.ConvertRequest) (*pb.ConvertResponse, error) {
	if _, err := os.Stat(req.Path); os.IsNotExist(err) {
		return nil, errors.New("docx file does not exist")
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Hey this is error")
		return nil, err
	}
	pdfPath := filepath.Join(cwd, fmt.Sprintf("%s.pdf", strings.TrimSuffix(filepath.Base(req.Path), ".docx")))

	args := []string{"--headless", "--convert-to", "pdf", "--outdir", ".", req.Path}
	cmd := exec.Command("libreoffice", args...)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return nil, errors.New("pdf file was not created")
	}

	return &pb.ConvertResponse{Path: pdfPath}, nil
}

func main() {

	lis, err := net.Listen("tcp", "localhost:9001")
	if err != nil {
		log.Fatalf("failed to listen: %+v", err)
	}

	s := grpc.NewServer()
	pb.RegisterConvertServiceServer(s, &server{})

	fmt.Println("Listen RPC server...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %+v", err)
	}
}
