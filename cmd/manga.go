package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/Nobodywinsbutme/mangahub/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var mangaCmd = &cobra.Command{
	Use:   "manga",
	Short: "Manga search commands using the gRPC service",
}

var mangaSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search manga through gRPC",
	Run: func(cmd *cobra.Command, args []string) {
		query, _ := cmd.Flags().GetString("query")
		client, closeFn := mustMangaClient()
		defer closeFn()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.SearchManga(ctx, &pb.SearchMangaRequest{Query: query})
		if err != nil {
			log.Fatalf("gRPC search failed: %v", err)
		}

		fmt.Printf("Found %d manga\n", len(resp.Results))
		for _, manga := range resp.Results {
			fmt.Printf("- %s | %s | %s | %d chapters | %s\n", manga.Id, manga.Title, manga.Author, manga.TotalChapters, manga.Status)
		}
	},
}

var mangaGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get manga details through gRPC",
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")
		client, closeFn := mustMangaClient()
		defer closeFn()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		manga, err := client.GetManga(ctx, &pb.GetMangaRequest{MangaId: id})
		if err != nil {
			log.Fatalf("gRPC get failed: %v", err)
		}

		fmt.Printf("%s\nAuthor: %s\nStatus: %s\nChapters: %d\n", manga.Title, manga.Author, manga.Status, manga.TotalChapters)
	},
}

var mangaProgressCmd = &cobra.Command{
	Use:   "grpc-progress",
	Short: "Update progress through gRPC",
	Run: func(cmd *cobra.Command, args []string) {
		userID, _ := cmd.Flags().GetString("user-id")
		mangaID, _ := cmd.Flags().GetString("manga-id")
		chapter, _ := cmd.Flags().GetInt("chapter")
		status, _ := cmd.Flags().GetString("status")

		client, closeFn := mustMangaClient()
		defer closeFn()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.UpdateProgress(ctx, &pb.UpdateProgressRequest{
			UserId:         userID,
			MangaId:        mangaID,
			CurrentChapter: int32(chapter),
			Status:         status,
		})
		if err != nil {
			log.Fatalf("gRPC progress update failed: %v", err)
		}

		fmt.Println(resp.Message)
	},
}

func mustMangaClient() (pb.MangaServiceClient, func()) {
	conn, err := grpc.NewClient("localhost:9092", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC service: %v", err)
	}
	return pb.NewMangaServiceClient(conn), func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close gRPC connection: %v", err)
		}
	}
}

func init() {
	mangaSearchCmd.Flags().String("query", "", "Search keyword")
	mangaSearchCmd.MarkFlagRequired("query")

	mangaGetCmd.Flags().String("id", "", "Manga ID")
	mangaGetCmd.MarkFlagRequired("id")

	mangaProgressCmd.Flags().String("user-id", "usr_123", "User ID")
	mangaProgressCmd.Flags().String("manga-id", "", "Manga ID")
	mangaProgressCmd.Flags().Int("chapter", 0, "Current chapter")
	mangaProgressCmd.Flags().String("status", "reading", "Reading status")
	mangaProgressCmd.MarkFlagRequired("manga-id")
	mangaProgressCmd.MarkFlagRequired("chapter")

	mangaCmd.AddCommand(mangaSearchCmd, mangaGetCmd, mangaProgressCmd)
	rootCmd.AddCommand(mangaCmd)
}
