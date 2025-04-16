package glover

import (
	"errors"
	"fmt"
	"log"
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
	log.Printf("rootCmd init")
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.glover.yaml)")
}

func initConfig() {
	log.Printf("initConfig start")

	if cfgFile != "" {
		log.Printf("Using config file: %s\n", cfgFile)
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		log.Printf("Using default config file\n")
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
			log.Printf("Error reading config file: %s\n", err)
			os.Exit(1)
		}
	}

	log.Printf("initConfig done")
}

func createExampleConfig() {
	exampleConfig := `
port = 8080
`
	configPath := "./.glover.toml"

	err := os.WriteFile(configPath, []byte(exampleConfig), 0o644)
	if err != nil {
		log.Printf("Error creating example config file: %s\n", err)
		os.Exit(1)
	}

	log.Printf("Example config file created at %s\n", configPath)
}

// set values to the PFlag variables from config, if they are set. Priority is still given to explicitly provided CLI flags.
func bindFlags(cmd *cobra.Command, _ []string) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// If using camelCase in the config file, replace hyphens with a camelCased string.
		// Since viper does case-insensitive comparisons, we don't need to bother fixing the case, and only need to remove the hyphens.
		configName := strings.ReplaceAll(f.Name, "-", "")

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && viper.IsSet(configName) {
			log.Printf("Binding flag %s to config value %s\n", f.Name, configName)
			val := viper.Get(configName)

			err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				log.Printf("Error setting flag %s: %s\n", f.Name, err)
				panic(err)
			}

			log.Printf("Flag '%s' set to config value %v\n", f.Name, val)
		}
	})
}
