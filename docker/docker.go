package docker

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/mengelbart/bwe-test-runner/common"
)

type testcase struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	URL         string            `json:"url"`
	ComposeFile string            `json:"compose_file"`
	Duration    common.Duration   `json:"duration"`
	RouterMap   map[string]string `json:"router_map"`
	LeftRouter  []tcPhase         `json:"left_router"`
	RightRouter []tcPhase         `json:"right_router"`
}

type implementations []implementation

type implementation struct {
	Name     string   `json:"name"`
	Sender   endpoint `json:"sender"`
	Receiver endpoint `json:"receiver"`
}

type endpoint struct {
	Image string
	Args  string
}

var (
	//go:embed testcases.json
	availableTestcasesJSON string
	availableTestcases     map[string]testcase

	//go:embed implementations.json
	availableImplementationsJSON string
	availableImplementations     map[string]implementation
)

func init() {
	mustParseConfig(availableTestcasesJSON, &availableTestcases)
	mustParseConfig(availableImplementationsJSON, &availableImplementations)
}

type Basic struct {
	Date int64
	testcase
	implementations
}

func NewBasic(date int64, testcaseNames string, implementationNames []string) common.Runner {
	is := implementations{}
	for _, in := range implementationNames {
		is = append(is, availableImplementations[in])
	}
	b := &Basic{
		Date:            date,
		testcase:        availableTestcases[testcaseNames],
		implementations: is,
	}
	return b
}

func (d *Basic) dumpConfig() error {
	impls := []common.Implementation{}
	for j, i := range d.implementations {
		impls = append(impls, common.Implementation{
			Name:   i.Name,
			Router: d.RouterMap[string('a'+j)],
			Source: string('a' + j),
		})
	}
	config := common.Config{
		Date:            d.Date,
		Implementations: impls,
		Scenario: common.Scenario{
			Name:        d.Name,
			Description: d.Description,
			URL:         d.URL,
		},
	}
	bs, err := json.Marshal(config)
	if err != nil {
		return err
	}
	f, err := os.Create("output/config.json")
	if err != nil {
		return err
	}
	_, err = f.Write(bs)
	return err
}

func (d *Basic) Run() error {
	const (
		leftRouterLogFile  = "output/leftrouter.log"
		rightRouterLogFile = "output/rightrouter.log"
	)
	for i := range d.implementations {
		for _, path := range []string{
			"send_log", "receive_log", "output",
		} {
			letter := string('a' + i)
			dir := fmt.Sprintf("output/%v/%v", letter, path)
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}
		}
	}
	if err := d.dumpConfig(); err != nil {
		return err
	}

	upCMD := exec.Command(
		"docker-compose", "-f", d.ComposeFile,
		"up", "--force-recreate",
	)
	upCMD.Stdout = os.Stdout
	upCMD.Stderr = os.Stderr

	// Use host env
	upCMD.Env = os.Environ()

	for l, i := range d.implementations {
		letter := string('A' + l)
		upCMD.Env = append(upCMD.Env, fmt.Sprintf("SENDER_%v=%v", letter, i.Sender.Image))
		upCMD.Env = append(upCMD.Env, fmt.Sprintf("SENDER_%v_ARGS=%v", letter, i.Sender.Args))
		upCMD.Env = append(upCMD.Env, fmt.Sprintf("RECEIVER_%v=%v", letter, i.Receiver.Image))
		upCMD.Env = append(upCMD.Env, fmt.Sprintf("RECEIVER_%v_ARGS=%v", letter, i.Receiver.Args))
	}
	if err := upCMD.Start(); err != nil {
		return err
	}

	defer func() {
		downCMD := exec.Command("docker-compose", "-f", d.ComposeFile, "down")
		//downCMD.Stdout = os.Stdout
		//downCMD.Stderr = os.Stderr

		// Use host env
		downCMD.Env = os.Environ()
		if err1 := downCMD.Run(); err1 != nil {
			log.Printf("failed to shutdown docker compose setup: %v\n", err1)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	leftRouterLog, err := os.Create(leftRouterLogFile)
	if err != nil {
		return err
	}
	rightRouterLog, err := os.Create(rightRouterLogFile)
	if err != nil {
		return err
	}

	tsCh := make(chan *TrafficShaper)
	errCh := make(chan error)
	go func() {
		ts, err1 := newTrafficShaper(ctx, "/leftrouter", d.LeftRouter, leftRouterLog)
		if err1 != nil {
			errCh <- err1
		}
		tsCh <- ts
	}()
	go func() {
		ts, err1 := newTrafficShaper(ctx, "/rightrouter", d.RightRouter, rightRouterLog)
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
		case <-time.After(d.Duration.Duration):
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
