package cmd

import (
	"errors"
	"fmt"
	"os"

	. "github.com/hajimehoshi/wasmserve/pkg"

	"github.com/spf13/cobra"
)

var flagConf string
var flagHTTP string
var flagTags string
var flagAllowOrigin string
var flagOverlay string

var rootCmd = &cobra.Command{
	Use:   "wasmserve",
	Short: "wasmserve is a web assembly server for golang",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(initCmd)

	buildCmd.Flags().StringVarP(&flagConf, "config", "c", DefaultTomlFile, "Which config file to use")
	watchCmd.Flags().StringVarP(&flagConf, "config", "c", DefaultTomlFile, "Which config file to use")
	runCmd.Flags().StringVarP(&flagConf, "config", "c", DefaultTomlFile, "Which config file to use")

	// TODO Test http
	runCmd.Flags().StringVarP(&flagHTTP, "http", "p", DefaultHttp, "HTTP bind address to serve")
	// TODO Test tags
	runCmd.Flags().StringVarP(&flagTags, "tags", "t", DefaultTags, "Build tags")
	// TODO Test allow origin
	runCmd.Flags().StringVarP(&flagAllowOrigin, "allow-origin", "a", DefaultAllowOrigin, "Allow specified origin (or * for all origins) to make requests to this server")
	// TODO Test overlay
	runCmd.Flags().StringVarP(&flagOverlay, "overlay", "o", DefaultOverlay, "Overwrite source files with a JSON file (see https://pkg.go.dev/cmd/go for more details)")
}

// returns whether the user uses config .toml or not
func useConfig() bool {
	if _, err := os.Stat(flagConf); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func initConf() error {
	if useConfig() {
		c, err := ReadConfig(flagConf)
		if err != nil {
			return err
		}
		*Config = *c
	} else {
		*Config = DefaultConfig()
	}
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
