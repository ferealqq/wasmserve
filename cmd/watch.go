package cmd

import (
	"log"

	"github.com/cosmtrek/air/runner"
	. "github.com/hajimehoshi/wasmserve/pkg"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch changes in the current project",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := ReadConfig(*flagAirConf)
		*Config = *c
		if err != nil {
			log.Fatal(err)
			return
		}

		e, err := runner.NewEngine(*flagAirConf, true)
		if err != nil {
			log.Fatal(err.Error())
		}
		e.Run()
		// FIXME Stop the process
	},
}
