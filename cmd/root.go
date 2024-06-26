package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/oasisprotocol/cli/cmd/inspect"
	"github.com/oasisprotocol/cli/config"
	"github.com/oasisprotocol/cli/version"
	_ "github.com/oasisprotocol/cli/wallet/file"   // Register file wallet backend.
	_ "github.com/oasisprotocol/cli/wallet/ledger" // Register ledger wallet backend.
)

const (
	defaultMarker = " (*)"
)

var (
	cfgFile string
    argYes bool
    argDesc string
    argSymbol string
    argExponent uint8
    argEd25519Priv string

	rootCmd = &cobra.Command{
		Use:     "hela",
		Short:   "CLI for interacting with the HELA network",
		Version: version.Software,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func initVersions() {
	cobra.AddTemplateFunc("toolchain", func() interface{} { return version.Toolchain })
	cobra.AddTemplateFunc("sdk", func() interface{} { return version.GetOasisSDKVersion() })
	cobra.AddTemplateFunc("core", func() interface{} { return version.GetOasisCoreVersion() })

	rootCmd.SetVersionTemplate(`Software version: {{.Version}}
HELA SDK version: {{ sdk }}
HELA Core version: {{ core }}
Go toolchain version: {{ toolchain }}
`)
}

func initConfig() {
	v := viper.New()

	if cfgFile != "" {
		// Use config file from the flag.
		v.SetConfigFile(cfgFile)
	} else {
		const configFilename = "cli.toml"
		configDir := config.Directory()
		configPath := filepath.Join(configDir, configFilename)

		v.AddConfigPath(configDir)
		v.SetConfigType("toml")
		v.SetConfigName(configFilename)

		// Ensure the configuration file exists.
		_ = os.MkdirAll(configDir, 0o700)
		if _, err := os.Stat(configPath); errors.Is(err, fs.ErrNotExist) {
			if _, err := os.Create(configPath); err != nil {
				cobra.CheckErr(fmt.Errorf("failed to create configuration file: %w", err))
			}

			// Populate the initial configuration file with defaults.
			config.ResetDefaults()
			_ = config.Save(v)
		}
	}

	_ = v.ReadInConfig()

	// Load and validate global configuration.
	err := config.Load(v)
	cobra.CheckErr(err)
	err = config.Global().Validate()
	cobra.CheckErr(err)
}

func init() {
	initVersions()

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file to use")

	rootCmd.PersistentFlags().BoolVar(&argYes          , "yes",         false, "always yes to confirm")
	rootCmd.PersistentFlags().StringVar(&argDesc       , "desc",         "",   "description")
	rootCmd.PersistentFlags().StringVar(&argSymbol     , "symbol",       "",   "token symbol")
	rootCmd.PersistentFlags().StringVar(&argEd25519Priv, "ed25519-priv", "",   "ed25519-raw private key")
	rootCmd.PersistentFlags().Uint8Var(&argExponent    , "exponent",     0,    "token exponent")

	rootCmd.AddCommand(networkCmd)
	rootCmd.AddCommand(paratimeCmd)
	rootCmd.AddCommand(walletCmd)
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(addressBookCmd)
	rootCmd.AddCommand(contractsCmd)
	rootCmd.AddCommand(inspect.Cmd)
	rootCmd.AddCommand(txCmd)
	rootCmd.AddCommand(managestCmd)
}
