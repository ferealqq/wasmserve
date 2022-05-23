package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"

	. "github.com/hajimehoshi/wasmserve/pkg"

	"github.com/cosmtrek/air/runner"
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
		log.Println("initConf")
		if err := initConf(); err != nil {
			log.Fatal(err)
			return
		}
		log.Println("init success")
		if flagConf != DefaultTomlFile {
			log.Println("Remember to change [build] cmd, bin and full_bin args to use custom config file")
		}
		log.Printf("flagConf => %s\n", flagConf)
		e, err := runner.NewEngine(flagConf, true)
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
		log.Println("running engine")
		e.Run()
	},
}
