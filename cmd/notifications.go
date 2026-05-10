package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/Nobodywinsbutme/mangahub/internal/udp_client"
)

var notificationsCmd = &cobra.Command{
	Use:   "notifications",
	Short: "Notification commands",
}

var notificationsListenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen for chapter release notifications",
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")

		client, err := udp_client.Connect("localhost", "9091")
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer client.Close()

		if err := client.Register(username); err != nil {
			log.Fatalf("Failed to register: %v", err)
		}

		fmt.Printf("Listening for notifications (%s)...\n", username)

		err = client.ListenForNotifications(func(msg string) error {
			fmt.Println("🔔", msg)
			return nil
		})
		if err != nil {
			log.Fatalf("Listen error: %v", err)
		}
	},
}

func init() {
	notificationsListenCmd.Flags().String("username", "", "Username")
	notificationsListenCmd.MarkFlagRequired("username")

	notificationsCmd.AddCommand(notificationsListenCmd)
	rootCmd.AddCommand(notificationsCmd)
	notificationsSendCmd.Flags().String("title", "", "Manga title")
	notificationsSendCmd.Flags().Int("chapter", 0, "Chapter number")
	notificationsSendCmd.MarkFlagRequired("title")
	notificationsSendCmd.MarkFlagRequired("chapter")

	notificationsCmd.AddCommand(notificationsSendCmd)
}

var notificationsSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a notification to all listeners",
	Run: func(cmd *cobra.Command, args []string) {
		title, _ := cmd.Flags().GetString("title")
		chapter, _ := cmd.Flags().GetInt("chapter")

		client, err := udp_client.Connect("localhost", "9091")
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer client.Close()

		if err := client.SendNotification(title, chapter); err != nil {
			log.Fatalf("Failed to send notification: %v", err)
		}

		fmt.Printf("✓ Notification sent: %s - Chapter %d\n", title, chapter)
	},
}
