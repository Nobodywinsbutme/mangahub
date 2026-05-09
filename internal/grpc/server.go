package grpc

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	"github.com/Nobodywinsbutme/mangahub/internal/database"
	pb "github.com/Nobodywinsbutme/mangahub/proto" // Adjust path to match your module

	"google.golang.org/grpc"
)

type MangaServer struct {
	pb.UnimplementedMangaServiceServer
}

func (s *MangaServer) GetManga(ctx context.Context, req *pb.GetMangaRequest) (*pb.GetMangaResponse, error) {
	var resp pb.GetMangaResponse

	// Assuming your table is named 'mangas'
	query := `SELECT id, title, author, total_chapters, status FROM mangas WHERE id = ?`
	err := database.DB.QueryRowContext(ctx, query, req.MangaId).Scan(
		&resp.Id, &resp.Title, &resp.Author, &resp.TotalChapters, &resp.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("manga not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	return &resp, nil
}

func (s *MangaServer) SearchManga(ctx context.Context, req *pb.SearchMangaRequest) (*pb.SearchMangaResponse, error) {
	query := `SELECT id, title, author, total_chapters, status FROM mangas WHERE title LIKE ?`

	// Add wildcards for fuzzy searching
	searchTerm := "%" + req.Query + "%"
	rows, err := database.DB.QueryContext(ctx, query, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	defer rows.Close()

	var results []*pb.GetMangaResponse
	for rows.Next() {
		var manga pb.GetMangaResponse
		if err := rows.Scan(&manga.Id, &manga.Title, &manga.Author, &manga.TotalChapters, &manga.Status); err != nil {
			continue // Skip problematic rows
		}
		results = append(results, &manga)
	}

	return &pb.SearchMangaResponse{Results: results}, nil
}

func (s *MangaServer) UpdateProgress(ctx context.Context, req *pb.UpdateProgressRequest) (*pb.UpdateProgressResponse, error) {
	// Using SQLite's UPSERT syntax (ON CONFLICT DO UPDATE)
	query := `
		INSERT INTO user_progress (user_id, manga_id, current_chapter, status) 
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, manga_id) 
		DO UPDATE SET current_chapter = excluded.current_chapter, status = excluded.status
	`

	_, err := database.DB.ExecContext(ctx, query, req.UserId, req.MangaId, req.CurrentChapter, req.Status)
	if err != nil {
		return &pb.UpdateProgressResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update progress: %v", err),
		}, nil
	}

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
