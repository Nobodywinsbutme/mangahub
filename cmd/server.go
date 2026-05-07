package cmd

import (
	"log"

	"github.com/Nobodywinsbutme/mangahub/internal/database"
	"github.com/Nobodywinsbutme/mangahub/internal/http_server"
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

		log.Println("Starting MangaHub servers...")

		// 2. For now, just start HTTP. We'll add the other 4 here in Phase 2.
		// Each server will be launched in its own goroutine so they run concurrently.
		http_server.Start("8080") // This blocks — move to goroutine in Phase 2
	},
}

func init() {
	// Wire subcommands to parent
	serverCmd.AddCommand(serverStartCmd)
	// Wire server command to root
	rootCmd.AddCommand(serverCmd)
}
