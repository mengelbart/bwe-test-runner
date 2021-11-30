package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	_ "embed"

	"github.com/spf13/cobra"
)

// tsTimeout is the time to wait for containers to spin up before timing out
const tsTimeout = 5 * time.Minute

// embedded static data
var (
	//go:embed testcases.json
	defaultTestcases string

	//go:embed implementations.json
	defaultImplementations string
)

// errors
var (
	errTrafficShaperTimeout  = errors.New("traffic shaper timed out while waiting for containers to spin up")
	errUnknownScenario       = errors.New("unknown scenario")
	errUnknownImplementation = errors.New("unknown implementation")
)

// flags
var (
	scenarioFlag       string
	implementationFlag string

	tc testcases
	is implementations
)

func testcaseNames(m testcases) []string {
	res := []string{}
	for k := range m {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

func implementationNames(i implementations) []string {
	res := []string{}
	for k := range i {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

func init() {
	rootCmd.AddCommand(runCmd)

	err := json.Unmarshal([]byte(defaultTestcases), &tc)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal([]byte(defaultImplementations), &is)
	if err != nil {
		panic(err)
	}

	runCmd.Flags().StringVarP(&scenarioFlag, "scenario", "s", "1", fmt.Sprintf("Test case scenario to run (options: %v)", strings.Join(testcaseNames(tc), ", ")))
	runCmd.Flags().StringVarP(&implementationFlag, "implementation", "i", "pion", fmt.Sprintf("Implementation to run (options: %v)", strings.Join(implementationNames(is), ", ")))
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

	t, ok := tc[scenarioFlag]
	if !ok {
		return errUnknownScenario
	}
	i, ok := is[implementationFlag]
	if !ok {
		return errUnknownImplementation
	}
	return runTestcase(t, i)
}

type implementations map[string]implementation

type endpoint struct {
	Image string `json:"image"`
	Args  string `json:"args"`
}

type implementation struct {
	Sender   endpoint `json:"Sender"`
	Receiver endpoint `json:"Receiver"`
}

func runTestcase(tc testcase, i implementation) error {
	upCMD := exec.Command(
		"docker-compose", "-f", tc.DCFile,
		"up", "--abort-on-container-exit", "--force-recreate",
	)
	upCMD.Stdout = os.Stdout
	upCMD.Stderr = os.Stderr

	// Use host env
	upCMD.Env = os.Environ()
	for k, v := range map[string]string{
		"SENDER_A":        i.Sender.Image,
		"RECEIVER_A":      i.Receiver.Image,
		"SENDER_A_ARGS":   i.Sender.Args,
		"RECEIVER_A_ARGS": i.Receiver.Args,

		"SENDER_B":        i.Sender.Image,
		"RECEIVER_B":      i.Receiver.Image,
		"SENDER_B_ARGS":   i.Sender.Args,
		"RECEIVER_B_ARGS": i.Receiver.Args,
	} {
		upCMD.Env = append(upCMD.Env, fmt.Sprintf("%v=%v", k, v))
	}
	if err := upCMD.Start(); err != nil {
		return err
	}

	defer func() {
		downCMD := exec.Command("docker-compose", "-f", tc.DCFile, "down")
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

	leftRouterLog, err := os.Create("leftrouter.log")
	if err != nil {
		return err
	}
	defer leftRouterLog.Close()
	rightRouterLog, err := os.Create("rightrouter.log")
	if err != nil {
		return err
	}
	defer rightRouterLog.Close()

	tsCh := make(chan *TrafficShaper)
	errCh := make(chan error)
	go func() {
		ts, err1 := newTrafficShaper(ctx, "/leftrouter", tc.Leftrouter, leftRouterLog)
		if err1 != nil {
			errCh <- err1
		}
		tsCh <- ts
	}()
	go func() {
		ts, err1 := newTrafficShaper(ctx, "/rightrouter", tc.Rightrouter, rightRouterLog)
		if err1 != nil {
			errCh <- err1
		}
		tsCh <- ts
	}()

	tss := []*TrafficShaper{}
	for i := 0; i < 2; i++ {
		select {
		case err = <-errCh:
			return err
		case ts := <-tsCh:
			tss = append(tss, ts)
		}
	}

	for _, ts := range tss {
		go func(ts *TrafficShaper) {
			err = ts.run(ctx)
			if err != nil {
				errCh <- err
			}
		}(ts)
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan error, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
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

	err = <-done
	return err
}
