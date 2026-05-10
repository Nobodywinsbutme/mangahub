package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/Nobodywinsbutme/mangahub/internal/ws_client"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Real-time WebSocket chat",
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")

		client, err := ws_client.Connect("localhost", "9093", username)
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer client.Close()

		fmt.Println("Connected. Type messages and press Enter.")

		// Listen for messages
		go func() {
			err := client.Receive(func(msg ws_client.Message) error {
				fmt.Printf("[%s] %s\n", msg.Username, msg.Text)
				return nil
			})
			if err != nil {
				log.Fatalf("Receive error: %v", err)
			}
		}()

		// Read stdin and send
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			if err := client.Send(text); err != nil {
				log.Fatalf("Send error: %v", err)
			}
		}
	},
}

func init() {
	chatCmd.Flags().String("username", "", "Username")
	chatCmd.MarkFlagRequired("username")

	rootCmd.AddCommand(chatCmd)
}
