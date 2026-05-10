package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/Nobodywinsbutme/mangahub/internal/tcp_client"
)

var progressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Reading progress commands",
}

var progressUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update reading progress",
	Run: func(cmd *cobra.Command, args []string) {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		chapter, _ := cmd.Flags().GetInt("chapter")

		// TODO: Replace with real user ID from JWT/session later
		userID := "usr_123"

		// Send progress to TCP sync server
		client, err := tcp_client.Connect("localhost", "9090")
		if err != nil {
			log.Fatalf("Failed to connect to sync server: %v", err)
		}
		defer client.Close()

		if err := client.SendProgress(userID, mangaID, chapter); err != nil {
			log.Fatalf("Failed to send progress: %v", err)
		}

		fmt.Printf("✓ Progress sent: %s - Chapter %d\n", mangaID, chapter)
	},
}

func init() {
	progressUpdateCmd.Flags().String("manga-id", "", "Manga ID")
	progressUpdateCmd.Flags().Int("chapter", 0, "Current chapter")
	progressUpdateCmd.MarkFlagRequired("manga-id")
	progressUpdateCmd.MarkFlagRequired("chapter")

	progressCmd.AddCommand(progressUpdateCmd)
	rootCmd.AddCommand(progressCmd)
}
