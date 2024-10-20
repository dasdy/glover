/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"glover/db"
	"glover/server"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("show called")

		log.Printf("Output file: %s\n", storagePath)
		storage, err := db.ConnectDB(storagePath)
		if err != nil {
			return fmt.Errorf("Could not open %s as sqlite file: %w", storagePath, err)
		}
		defer storage.Close()
		log.Printf("Runnint interface on port %d\n", port)
		err = http.ListenAndServe(
			fmt.Sprintf(":%d", port),
			server.BuildServer(storage))
		if err != nil {
			log.Fatalf("Could not run server: %s", err)
		}

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
}
