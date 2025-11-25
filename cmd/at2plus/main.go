package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "at2plus",
	Short: "AirTouch 2+ Control CLI",
	Long:  `A command line interface for controlling AirTouch 2+ air conditioner controllers.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
