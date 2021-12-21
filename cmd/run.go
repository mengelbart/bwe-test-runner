package cmd

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mengelbart/bwe-test-runner/common"
	"github.com/mengelbart/bwe-test-runner/docker"
	"github.com/spf13/cobra"
)

var (
	runDate            int64
	runnerFlag         string
	scenarioFlag       string
	implementationFlag string
)

var runnersMap = map[string]common.RunnerFactory{
	"docker": docker.NewBasic,
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "execute test run",
	Run: func(_ *cobra.Command, _ []string) {
		if err := run(); err != nil {
			log.Fatal(err)
		}
	},
}

func runners() []string {
	keys := []string{}
	for k := range runnersMap {
		keys = append(keys, k)
	}
	return keys
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().Int64VarP(&runDate, "date", "d", time.Now().Unix(), "Unix Timestamp in seconds since epoch")
	runCmd.Flags().StringVarP(&runnerFlag, "runner", "r", runners()[0], fmt.Sprintf("Test case scenario to run (options: %v)", strings.Join(runners(), ", ")))
	runCmd.Flags().StringVarP(&scenarioFlag, "scenario", "s", "1", "Scenario to run")
	runCmd.Flags().StringVarP(&implementationFlag, "implementation", "i", "pion-gcc", "Implementation to run")
}

var errInvalidRunner = errors.New("invalid runner")

func run() error {
	runnerFactory, ok := runnersMap[runnerFlag]
	if !ok {
		return errInvalidRunner
	}
	runner := runnerFactory(runDate, scenarioFlag, implementationFlag)
	return runner.Run()
}
