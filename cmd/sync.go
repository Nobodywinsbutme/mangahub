package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/Nobodywinsbutme/mangahub/internal/tcp_client"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync features",
}

var syncConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to real-time sync server",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := tcp_client.Connect("localhost", "9090")
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer client.Close()

		fmt.Println("Connected to sync server. Listening for updates...")

		err = tcp_client.ListenForUpdates(client, func(msg string) error {
			fmt.Println("Update:", msg)
			return nil
		})
		if err != nil {
			log.Fatalf("Listen error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.AddCommand(syncConnectCmd)
}
