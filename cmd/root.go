package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mangahub",
	Short: "MangaHub - Your manga tracking system",
	Long: `MangaHub is a CLI application for tracking manga and comics across multiple
communication protocols (HTTP, TCP, UDP, WebSocket, gRPC).`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
