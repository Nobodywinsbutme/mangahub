package cmd

import (
	"log"

	"github.com/Nobodywinsbutme/mangahub/internal/database"
	"github.com/Nobodywinsbutme/mangahub/internal/grpc"
	"github.com/Nobodywinsbutme/mangahub/internal/http_server"
	"github.com/Nobodywinsbutme/mangahub/internal/tcp"
	"github.com/Nobodywinsbutme/mangahub/internal/udp"
	"github.com/Nobodywinsbutme/mangahub/internal/websocket"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage MangaHub server components",
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start all MangaHub server components",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Initialize DB first — everything depends on it
		if err := database.Init("./mangahub.db"); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}

		log.Println("Starting MangaHub Multi-Protocol Backend...")

		// 2. Launch servers as independent goroutines
		go http_server.Start("8080")
		go tcp.Start("9090")
		go udp.Start("9091")
		go grpc.Start("9092")
		go websocket.Start("9093")

		log.Println("✅ All 5 servers are up and running! Press Ctrl+C to stop.")

		// 3. Block the main thread indefinitely so background goroutines stay alive
		select {}
	},
}

func init() {
	// Wire subcommands to parent
	serverCmd.AddCommand(serverStartCmd)
	// Wire server command to root
	rootCmd.AddCommand(serverCmd)
}
