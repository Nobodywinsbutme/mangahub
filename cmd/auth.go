package cmd

import (
	"fmt"
	"log"
	"syscall"

	"github.com/Nobodywinsbutme/mangahub/internal/auth"
	"github.com/Nobodywinsbutme/mangahub/internal/database"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new account",
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")
		email, _ := cmd.Flags().GetString("email")

		// Read password securely (no terminal echo)
		fmt.Print("Password: ")
		passBytes, _ := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		password := string(passBytes)

		database.Init("./mangahub.db")

		user, err := auth.RegisterUser(username, email, password)
		if err != nil {
			log.Fatalf("✗ Registration failed: %v", err)
		}
		fmt.Printf("✓ Account created! User ID: %s\n", user.ID)
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your account",
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")

		fmt.Print("Password: ")
		passBytes, _ := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()

		database.Init("./mangahub.db")

		token, user, err := auth.LoginUser(username, string(passBytes))
		if err != nil {
			log.Fatalf("✗ Login failed: %v", err)
		}
		fmt.Printf("✓ Welcome back, %s!\nToken: %s\n", user.Username, token)
	},
}

func init() {
	registerCmd.Flags().String("username", "", "Your username (required)")
	registerCmd.Flags().String("email", "", "Your email (required)")
	registerCmd.MarkFlagRequired("username")
	registerCmd.MarkFlagRequired("email")

	loginCmd.Flags().String("username", "", "Your username (required)")
	loginCmd.MarkFlagRequired("username")

	authCmd.AddCommand(registerCmd)
	authCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(authCmd)
}
