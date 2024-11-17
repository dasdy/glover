package glover

import (
	"fmt"
	"os"

	"github.com/dasdy/glover/db"
	"github.com/spf13/cobra"
)

// mergeCmd represents the merge command.
var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge two databases into one",
	Long:  `Given two log files, create a third one, which is just a union of input databases`,
	RunE: func(_ *cobra.Command, _ []string) error {
		inputs := make([]*db.SQLiteStorage, len(filenames))
		for i, fn := range filenames {
			// TODO: this automatically also counts all the combos, but we don't need it for merging.
			// maybe some simplified init can be created?
			store, err := db.NewStorageFromPath(fn, false)
			if err != nil {
				return err
			}
			inputs[i] = store
		}

		if _, err := os.Stat(storagePath); err == nil {
			return fmt.Errorf("output file %s already exists", storagePath)
		}

		output, err := db.NewStorageFromPath(storagePath, false)
		if err != nil {
			return err
		}

		err = db.Merge(inputs, output)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().StringSliceVarP(
		&filenames,
		"file",
		"f",
		[]string{},
		"List of filenames to merge data into",
	)

	mergeCmd.Flags().StringVarP(
		&storagePath,
		"out",
		"o",
		"./merged.sqlite",
		"Output path for statistics")
}
