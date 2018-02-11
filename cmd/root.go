package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rms1000watt/genni/generator"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var logLevel string
var generatorCfg generator.Config

var rootCmd = &cobra.Command{
	Use:   "genni",
	Short: "Genni generates go structs from go files",
	Long:  `Genni generates go structs from go files`,
	Run:   RootRun,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&generatorCfg.InFile, "in", "i", "", "Proto file to read in")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set log level (debug, info, warn, error, fatal)")

	setFlagsFromEnv(rootCmd)
	setPFlagsFromEnv(rootCmd)
}

func RootRun(cmd *cobra.Command, args []string) {
	configureLogging()
	generator.Generator(generatorCfg)
}

func setPFlagsFromEnv(cmd *cobra.Command) {
	// Courtesy of https://github.com/coreos/pkg/blob/master/flagutil/env.go
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		key := strings.ToUpper(strings.Replace(f.Name, "-", "_", -1))
		if val := os.Getenv(key); val != "" {
			if err := cmd.PersistentFlags().Set(f.Name, val); err != nil {
				fmt.Println("Failed setting flag from env:", err)
			}
		}
	})
}

func setFlagsFromEnv(cmd *cobra.Command) {
	// Courtesy of https://github.com/coreos/pkg/blob/master/flagutil/env.go
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		key := strings.ToUpper(strings.Replace(f.Name, "-", "_", -1))
		if val := os.Getenv(key); val != "" {
			if err := cmd.Flags().Set(f.Name, val); err != nil {
				fmt.Println("Failed setting flag from env:", err)
			}
		}
	})
}

func configureLogging() {
	if level, err := log.ParseLevel(logLevel); err != nil {
		log.Error("log-level argument malformed: ", logLevel, ": ", err)
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}
}
