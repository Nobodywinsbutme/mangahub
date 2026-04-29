package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

func main() {
	fmt.Println("Welcome to MangaHub CLI!")

	// Just a dummy reference so Go knows we are using these packages
	_ = gin.New()
	_ = &cobra.Command{}
}
