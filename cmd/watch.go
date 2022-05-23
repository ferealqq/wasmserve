package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/cosmtrek/air/runner"
	. "github.com/hajimehoshi/wasmserve/pkg"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch changes in the current project",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		const (
			exitCodeErr       = 1
			exitCodeInterrupt = 2
		)

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

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		defer func() {
			signal.Stop(signalChan)
			cancel()
		}()
		go func() {
			select {
			case <-signalChan:
				cancel()
				e.Stop()
			case <-ctx.Done():
			}
			<-signalChan
			os.Exit(exitCodeInterrupt)
		}()

		e.Run()
	},
}
