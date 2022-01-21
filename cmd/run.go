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
	testcaseFlag       string
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

	dockerTestCases := docker.TestCaseList()
	dockerImplementations := docker.ImplementationList()

	runCmd.Flags().Int64VarP(&runDate, "date", "d", time.Now().Unix(), "Unix Timestamp in seconds since epoch")
	runCmd.Flags().StringVarP(&implementationFlag, "implementation", "i", dockerImplementations[0], "Implementation to run")
	runCmd.Flags().StringVarP(&testcaseFlag, "testcase", "t", dockerTestCases[0], "Testcase to run")
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outputDir := path.Join("output/", fmt.Sprintf("%v", runDate), implementationFlag, "1")
	plotDir := path.Join("html/", fmt.Sprintf("%v", runDate), implementationFlag, "1")
	basetime := time.Now().Unix()
	errCh := make(chan error)

	tc := docker.TestCases[testcaseFlag]
	go func() {
		errCh <- tc.Run(ctx, implementationFlag, "input.y4m", outputDir)
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

	if err := tc.Plot("input.y4m", outputDir, plotDir, basetime); err != nil {
		return err
	}
	return nil
}
