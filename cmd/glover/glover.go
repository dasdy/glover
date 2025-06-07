package glover

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "glover",
	Short: "Track and visualize your keystrokes",
	Long: `Glover can help you visualize your usage of ZMK-backed keyboards.
It can build a database and visualize it as a heatmap, allowing you to take action
and optimize your layout as you see fit.`,
	PersistentPreRun: bindFlags,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	slog.Info("rootCmd init")
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.glover.yaml)")
}

func initConfig() {
	slog.Info("initConfig start")

	if cfgFile != "" {
		slog.Info("Using config file", "file", cfgFile)
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		slog.Info("Using default config file")
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".glover" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("toml")
		viper.SetConfigName(".glover")
	}
	// Set environment variable prefix
	viper.SetEnvPrefix("glover")
	viper.AutomaticEnv()

	// Read config
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// Config file not found, create an example config
			createExampleConfig()
		} else {
			// Other errors
			slog.Error("Error reading config file", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("initConfig done")
}

func createExampleConfig() {
	exampleConfig := `
port = 8080
`
	configPath := "./.glover.toml"

	err := os.WriteFile(configPath, []byte(exampleConfig), 0o644)
	if err != nil {
		slog.Error("Error creating example config file", "error", err)
		os.Exit(1)
	}

	slog.Info("Example config file created", "path", configPath)
}

// set values to the PFlag variables from config, if they are set. Priority is still given to explicitly provided CLI flags.
func bindFlags(cmd *cobra.Command, _ []string) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// If using camelCase in the config file, replace hyphens with a camelCased string.
		// Since viper does case-insensitive comparisons, we don't need to bother fixing the case, and only need to remove the hyphens.
		configName := strings.ReplaceAll(f.Name, "-", "")

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && viper.IsSet(configName) {
			slog.Info("binding flag", "key", f.Name, "config", configName)
			val := viper.Get(configName)

			err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				slog.Error("error setting flag", "key", f.Name, "error", err)
				panic(err)
			}

			slog.Info("setting flag value", "key", f.Name, "value", val)
		}
	})
}
