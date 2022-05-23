package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagHTTP        = flag.String("http", ":8080", "HTTP bind address to serve")
	flagTags        = flag.String("tags", "", "Build tags")
	flagAllowOrigin = flag.String("allow-origin", "", "Allow specified origin (or * for all origins) to make requests to this server")
	flagOverlay     = flag.String("overlay", "", "Overwrite source files with a JSON file (see https://pkg.go.dev/cmd/go for more details)")
	flagBuild       = flag.Bool("build", false, "Build tailwind & wasm")
	flagRun         = flag.Bool("run", false, "Run HTTP server")
	flagWatch       = flag.Bool("watch", false, "Watch file changes and serve http")
	flagAirConf     = flag.String("config", "wasmserve.toml", "Path to wasmserve.toml configuration")
	flagInit        = flag.Bool("init", false, "create initial configurations")
)

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
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
