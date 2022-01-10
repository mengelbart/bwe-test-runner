package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/mengelbart/bwe-test-runner/docker"
	"github.com/spf13/cobra"
)

var (
	runDate            int64
	implementationFlag string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "execute test run",
	Run: func(_ *cobra.Command, _ []string) {
		if err := run(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().Int64VarP(&runDate, "date", "d", time.Now().Unix(), "Unix Timestamp in seconds since epoch")
	runCmd.Flags().StringVarP(&implementationFlag, "implementation", "i", "pion-gcc", "Implementation to run")
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outputDir := path.Join("output/", fmt.Sprintf("%v", runDate), implementationFlag, "1")
	plotDir := path.Join("html/", fmt.Sprintf("%v", runDate), implementationFlag, "1")
	basetime := time.Now().Unix()
	errCh := make(chan error)
	go func() {
		errCh <- docker.Run(ctx, implementationFlag, outputDir)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		select {
		case sig := <-sigs:
			fmt.Printf("got signal %v, aborting\n", sig)
			cancel()
		case <-ctx.Done():
		}
	}()

	err := <-errCh
	if err != nil {
		return err
	}

	if err := docker.Plot(outputDir, plotDir, basetime); err != nil {
		return err
	}
	return nil
}
