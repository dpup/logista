package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dpup/logista/internal/formatter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version is set during build
var version = "dev"

// Default format template
const defaultFormat = "{{.timestamp | date}} {{.level}} {{.message}}"

// Config options
const (
	keyFormat       = "format"
	keyDateFormat   = "date_format"
	keyNoColors     = "no_colors"
	keyConfig       = "config"
	keyEnableSimple = "enable_simple_syntax"
)

// Initialize cobra command
var rootCmd = &cobra.Command{
	Use:   "logista",
	Short: "Logista is a tool for formatting JSON log streams",
	Long: `Logista is a CLI tool that accepts a stream of JSON log entries 
and formats them according to a specified template.`,
	RunE:    runLogista,
	Version: version,
}

var cfgFile string

func init() { //nolint:gochecknoinits // Required for cobra command initialization
	cobra.OnInitialize(initConfig)

	// Config file flag
	rootCmd.PersistentFlags().StringVar(&cfgFile, keyConfig, "", "config file (default is $HOME/.logista.yaml)")

	// Command line flags
	rootCmd.PersistentFlags().String(keyFormat, defaultFormat, "Format template")
	rootCmd.PersistentFlags().String(keyDateFormat, "2006-01-02 15:04:05", "Preferred date format for the date function")
	rootCmd.PersistentFlags().Bool(keyNoColors, false, "Disable colored output")
	rootCmd.PersistentFlags().Bool(keyEnableSimple, true, "Enable simple {field} syntax in templates")

	// Bind flags to viper
	if err := viper.BindPFlag(keyFormat, rootCmd.PersistentFlags().Lookup(keyFormat)); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding flag %s: %v\n", keyFormat, err)
	}
	if err := viper.BindPFlag(keyDateFormat, rootCmd.PersistentFlags().Lookup(keyDateFormat)); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding flag %s: %v\n", keyDateFormat, err)
	}
	if err := viper.BindPFlag(keyNoColors, rootCmd.PersistentFlags().Lookup(keyNoColors)); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding flag %s: %v\n", keyNoColors, err)
	}
	if err := viper.BindPFlag(keyEnableSimple, rootCmd.PersistentFlags().Lookup(keyEnableSimple)); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding flag %s: %v\n", keyEnableSimple, err)
	}

	// Set environment variable prefix
	viper.SetEnvPrefix("LOGISTA")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".logista" (without extension)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".logista")
	}

	// Read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// runLogista is the main function that processes the log stream
func runLogista(cmd *cobra.Command, args []string) error {
	// Apply options from configuration
	options := []formatter.FormatterOption{
		formatter.WithPreferredDateFormat(viper.GetString(keyDateFormat)),
	}

	// Add no-colors option if set
	if viper.GetBool(keyNoColors) {
		options = append(options, formatter.WithNoColors(true))
	}

	// Get format template from config
	formatTemplate := viper.GetString(keyFormat)

	// Create preprocessor options
	preprocessOptions := formatter.DefaultPreProcessTemplateOptions()
	preprocessOptions.EnableSimpleSyntax = viper.GetBool(keyEnableSimple)

	// Create the formatter with format template, preprocessor options, and formatter options
	tmplFormatter, err := formatter.NewTemplateFormatterWithOptions(formatTemplate, preprocessOptions, options...)
	if err != nil {
		return fmt.Errorf("invalid format template: %w", err)
	}

	return tmplFormatter.ProcessStream(os.Stdin, os.Stdout, tmplFormatter)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
