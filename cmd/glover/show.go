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
	"github.com/spf13/viper"
)

// showCmd represents the show command.
var showCmd = &cobra.Command{
	Use:              "show",
	Short:            "Show collected statistics",
	Long:             `Use log data collected by track command to show web interface with statistics.`,
	PersistentPreRun: bindFlags,
	RunE: func(_ *cobra.Command, _ []string) error {
		log.Printf("Config file: %s\n", viper.ConfigFileUsed())
		log.Printf("Config parameters: %v\n", viper.AllSettings())
		log.Printf("kmapfile: %s", viper.GetString("keymap-file"))
		log.Printf("Output file: %s\n", storagePath)

		storage, err := db.NewStorageFromPath(storagePath, true)
		if err != nil {
			return fmt.Errorf("could not open %s as sqlite file: %w", storagePath, err)
		}
		comboTracker, err := db.NewComboTrackerFromDB(storage)
		if err != nil {
			return fmt.Errorf("could not create combo tracker: %w", err)
		}
		neighborTracker, err := db.NewNeighborCounterFromDb(storage)
		if err != nil {
			return fmt.Errorf("could not create neighbor tracker: %w", err)
		}
		defer storage.Close()
		web.StartServer(port, storage, comboTracker, neighborTracker, keymapFile, dev)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)

	showCmd.Flags().IntVarP(&port, "port", "p", 9000,
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

	showCmd.Flags().StringVar(
		&keymapFile,
		"keymap-file",
		"glove80.keymap",
		"Path to the keymap file used for rendering the interface")
}
