package glover

import (
	"fmt"
	"log"
	"os"

	"github.com/dasdy/glover/db"
	"github.com/dasdy/glover/keylog"
	"github.com/dasdy/glover/keylog/ports"
	"github.com/dasdy/glover/web"
	"github.com/spf13/cobra"
)

// trackCmd represents the track command.
var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Connect to attached keyboard and log keypresses",
	Long: `Provide two paths to files to connect to, or leave empty to read from stdin.
		Will log keypresses to a sqlite file, and optionally run a web server to visualize the data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileCount := len(filenames)

		if fileCount != 2 && fileCount != 0 {
			return fmt.Errorf("expected exactly 0 or 2 files, got %d", len(filenames))
		}

		var ch <-chan string
		switch fileCount {
		case 0:
			names, err := ports.GetAvailableDevices()
			if err != nil {
				return err
			}
			log.Printf("Suggested devices: %+v ", names)
			log.Print("Will proceed to read from stdin...")

			ch = ports.ReadFile(os.Stdin)
		case 2:
			var closer func()
			var err error
			ch, closer, err = ports.OpenTwoFiles(filenames[0], filenames[1])
			defer closer()
			if err != nil {
				// Try suggesting devices
				names, errInner := ports.GetAvailableDevices()
				if errInner != nil {
					return fmt.Errorf("could not open file: %w; Could not suggest devices: %w", err, errInner)
				}

				if len(names) > 0 {
					return fmt.Errorf("error opening files: %w. Maybe try instead: %+v", err, names)
				} else {
					return fmt.Errorf("error opening files: %w. It does not seem like any keyboard is connected", err)
				}
			}
		}

		log.Printf("Output file: %s\n", storagePath)
		storage, err := db.ConnectDB(storagePath)
		if err != nil {
			return fmt.Errorf("could not open %s as sqlite file: %w", storagePath, err)
		}
		defer storage.Close()

		if !disableInterface {
			go web.StartServer(port, storage, dev)
		}

		log.Print("Main loop")
		keylog.KeyLogLoop(ch, storage, verbose)
		// In order for air to auto-restart, we need to return error code. It does not do this if
		// the code is 0
		return fmt.Errorf("I shall not accept just closure of file! Restart me!")
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
