package grpc

import (
	"context"
	"log"
	"net"

	pb "github.com/Nobodywinsbutme/mangahub/proto" // Adjust path to match your module

	"google.golang.org/grpc"
)

type MangaServer struct {
	pb.UnimplementedMangaServiceServer
}

func (s *MangaServer) GetManga(ctx context.Context, req *pb.GetMangaRequest) (*pb.GetMangaResponse, error) {
	// TODO: Integrate with internal/manga/service.go DB queries
	return &pb.GetMangaResponse{
		Id:            req.MangaId,
		Title:         "Sample Title",
		Author:        "Sample Author",
		TotalChapters: 100,
		Status:        "ongoing",
	}, nil
}

func (s *MangaServer) SearchManga(ctx context.Context, req *pb.SearchMangaRequest) (*pb.SearchMangaResponse, error) {
	// TODO: Integrate with internal/manga/service.go DB queries
	return &pb.SearchMangaResponse{}, nil
}

func (s *MangaServer) UpdateProgress(ctx context.Context, req *pb.UpdateProgressRequest) (*pb.UpdateProgressResponse, error) {
	// TODO: Integrate with internal/manga/service.go DB queries
	return &pb.UpdateProgressResponse{
		Success: true,
		Message: "Progress updated successfully",
	}, nil
}

func Start(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterMangaServiceServer(server, &MangaServer{})

	log.Printf("gRPC Internal Service listening on port %s", port)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("gRPC Server failed: %v", err)
	}
}
