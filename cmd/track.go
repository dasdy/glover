package cmd

import (
	"fmt"
	"glover/db"
	"glover/keylog"
	"glover/keylog/ports"
	"glover/server"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

// trackCmd represents the track command
var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("filenames: %+v\n", filenames)

		fileCount := len(filenames)

		if fileCount != 2 && fileCount != 0 {
			return fmt.Errorf("expected exactly 0 or 2 files, got %d", len(filenames))
		}

		if fileCount == 0 {
			names, err := ports.GetAvailableDevices()
			if err != nil {
				return err
			}

			log.Printf("Suggested devices: %+v ", names)

			log.Print("Will proceed to read from stdin...")
		}

		var ch <-chan string
		var closer func()
		var err error
		if fileCount == 2 {
			ch, closer, err = ports.OpenTwoFiles(filenames[0], filenames[1])
			defer closer()
			if err != nil {
				// Try suggesting devices
				names, errInner := ports.GetAvailableDevices()
				if errInner != nil {
					return fmt.Errorf("Could not open file: %w; Could not suggest devices: %w", err, errInner)
				}

				if len(names) > 0 {
					return fmt.Errorf("Error opening files: %w. Maybe try instead: %+v", err, names)
				} else {
					return fmt.Errorf("Error opening files: %w. It does not seem like any keyboard is connected...", err)
				}
			}
		} else {
			ch = ports.ReadFile(os.Stdin)
		}

		log.Printf("Output file: %s\n", storagePath)
		storage, err := db.ConnectDB(storagePath)
		if err != nil {
			return fmt.Errorf("Could not open %s as sqlite file: %w", storagePath, err)
		}
		defer storage.Close()

		if !disableInterface {
			log.Printf("Runnint interface on port %d\n", port)
			go func() {
				err := http.ListenAndServe(
					fmt.Sprintf(":%d", port),
					server.BuildServer(storage, dev))
				if err != nil {
					log.Fatalf("Could not run server: %s", err)
				}
			}()
		}

		log.Print("Main loop")
		keylog.KeyLogLoop(ch, storage, verbose)
		return nil
	},
}

var (
	filenames        []string
	storagePath      string
	port             int
	disableInterface bool
	verbose          bool
	dev              bool
)

func init() {
	rootCmd.AddCommand(trackCmd)
	trackCmd.Flags().StringSliceVarP(
		&filenames,
		"file",
		"f",
		[]string{},
		"List of filenames to get input from",
	)

	trackCmd.Flags().StringVarP(
		&storagePath,
		"out",
		"o",
		"./keypresses.sqlite",
		"Output path for statistics")

	trackCmd.Flags().IntVarP(
		&port, "port", "p", 3000,
		"Port on which server should be watching")

	trackCmd.Flags().BoolVar(&disableInterface,
		"no-interface",
		false,
		"If provided, no web server will be run with visualization")

	trackCmd.Flags().BoolVarP(&verbose,
		"verbose",
		"v",
		false,
		"If provided, debug output will be shown")

	trackCmd.Flags().BoolVar(&dev,
		"dev",
		false,
		"Enable developer mode")
}
