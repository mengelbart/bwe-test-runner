package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	_ "embed"

	"github.com/spf13/cobra"
)

//go:embed testcases.json
var defaultTestcases string

// tsTimeout is the time to wait for containers to spin up before timing out
const tsTimeout = 5 * time.Minute

var errTrafficShaperTimeout = errors.New("traffic shaper timed out while waiting for containers to spin up")

func init() {
	rootCmd.AddCommand(runCmd)
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

func run() error {

	var tc testcases
	err := json.Unmarshal([]byte(defaultTestcases), &tc)
	if err != nil {
		return err
	}

	return runTestcase(tc["5.1"])
}

func runTestcase(tc testcase) error {
	upCMD := exec.Command("docker-compose", "up", "--abort-on-container-exit", "--force-recreate")
	//upCMD.Stdout = os.Stdout
	//upCMD.Stderr = os.Stderr

	// Use host env
	upCMD.Env = os.Environ()
	if err := upCMD.Start(); err != nil {
		return err
	}

	defer func() {
		downCMD := exec.Command("docker-compose", "down")
		//downCMD.Stdout = os.Stdout
		//downCMD.Stderr = os.Stderr

		// Use host env
		downCMD.Env = os.Environ()
		if err := downCMD.Run(); err != nil {
			log.Printf("failed to shutdown docker compose setup: %v\n", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsCh := make(chan *TrafficShaper)
	errCh := make(chan error)
	go func() {
		ts, err := newTrafficShaper(ctx, "/leftrouter", tc.Leftrouter)
		if err != nil {
			errCh <- err
		}
		tsCh <- ts
	}()
	go func() {
		ts, err := newTrafficShaper(ctx, "/rightrouter", tc.Rightrouter)
		if err != nil {
			errCh <- err
		}
		tsCh <- ts
	}()

	tss := []*TrafficShaper{}
	for i := 0; i < 2; i++ {
		select {
		case err := <-errCh:
			return err
		case ts := <-tsCh:
			tss = append(tss, ts)
		}
	}

	for _, ts := range tss {
		go func(ts *TrafficShaper) {
			err := ts.run(ctx)
			if err != nil {
				errCh <- err
			}
		}(ts)
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan error, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		var err error
		select {
		case <-time.After(tc.Duration.Duration):
			log.Printf("testcase time over\n")

		case err1 := <-errCh:
			log.Printf("got error: %v, exiting\n", err)
			err = err1

		case sig := <-sigs:
			log.Printf("got signal %v, exiting\n", sig)
		}
		done <- err
	}()

	err := <-done
	return err
}
