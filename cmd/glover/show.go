/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package glover

import (
	"fmt"
	"log"

	"github.com/dasdy/glover/db"
	"github.com/dasdy/glover/web"
	"github.com/spf13/cobra"
)

// showCmd represents the show command.
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show collected statistics",
	Long:  `Use log data collected by track command to show web interface with statistics.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		log.Printf("Output file: %s\n", storagePath)
		storage, err := db.ConnectDB(storagePath)
		if err != nil {
			return fmt.Errorf("could not open %s as sqlite file: %w", storagePath, err)
		}
		defer storage.Close()
		web.StartServer(port, storage, dev)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)

	// Variables themselves are defined elsewhere
	showCmd.Flags().IntVarP(
		&port, "port", "p", 3000,
		"Port on which server should be watching")

	showCmd.Flags().StringVarP(
		&storagePath,
		"storage",
		"s",
		"./keypresses.sqlite",
		"Output path for statistics")

	showCmd.Flags().BoolVar(&dev,
		"dev",
		false,
		"Enable developer mode")
}
