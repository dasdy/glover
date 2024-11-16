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

func shouldTryConnect(names1 []string, names2 []string, autoconnect bool) bool {
	if !autoconnect || len(names1) != 2 || len(names2) != 2 {
		return false
	}

	for _, n1 := range names1 {
		for _, n2 := range names2 {
			if n1 == n2 {
				return false
			}
		}
	}
	return true
}

func GetInputsChannel(filenames []string, autoConnect bool) (<-chan string, func(), error) {
	fileCount := len(filenames)
	noop := func() {}

	if fileCount != 2 && fileCount != 0 {
		return nil, noop, fmt.Errorf("expected exactly 0 or 2 files, got %d", len(filenames))
	}

	switch fileCount {
	case 0:
		names, err := ports.GetAvailableDevices()
		if err != nil {
			return nil, noop, err
		}
		log.Printf("Suggested devices: %+v ", names)

		if autoConnect && len(names) == 2 {
			log.Print("Will proceed to autoconnect to devices")
			return GetInputsChannel(names, false)
		} else {
			log.Print("Will proceed to read from stdin...")
			return ports.ReadFile(os.Stdin), noop, nil
		}
	case 2:
		var closer func()
		var err error
		ch, closer, err := ports.OpenTwoFiles(filenames[0], filenames[1])

		if err == nil {
			return ch, closer, nil
		}

		// Try suggesting devices
		names, errInner := ports.GetAvailableDevices()
		if errInner != nil {
			return nil, closer, fmt.Errorf("could not open file: %w; Could not suggest devices: %w", err, errInner)
		}

		log.Printf("could not open provided files. Found candidates to connect instead: %+v", names)

		switch {
		case len(names) > 0 && shouldTryConnect(filenames, names, autoConnect):
			log.Print("autoconnect enabled. Trying to connect to candidates")
			return GetInputsChannel(names, false)

		case len(names) > 0:
			return nil, closer, fmt.Errorf("error opening files: %w.", err)

		default:
			return nil, closer, fmt.Errorf("error opening files: %w. It does not seem like any keyboard is connected", err)
		}
	}

	return nil, noop, fmt.Errorf("Strange count of devices provided: %d: %+v", len(filenames), filenames)
}

// trackCmd represents the track command.
var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Connect to attached keyboard and log keypresses",
	Long: `Provide two paths to files to connect to, or leave empty to read from stdin.
		Will log keypresses to a sqlite file, and optionally run a web server to visualize the data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ch, closer, err := GetInputsChannel(filenames, autoConnect)

		defer closer()
		if err != nil {
			return fmt.Errorf("could not open inputs channel: %w", err)
		}

		log.Printf("connected successfully. Output file: %s\n", storagePath)
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
	autoConnect      bool
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

	trackCmd.Flags().BoolVar(&autoConnect,
		"auto-connect",
		true,
		`If true, try connecting to available devices if provided ones do not work/nothing was provided. 
        If no devices can be found, use stdin. If auto-connect is false, always use stdin when no input devices are provided.`)
}
