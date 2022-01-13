package docker

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"time"
)

type TestCase struct {
	composeFileString string
	duration          time.Duration
	leftRouter        []tcPhase
	rightRouter       []tcPhase
	plotFunc          func(outputDir, plotDir string, basetime int64) error
}

func TestCaseList() []string {
	res := []string{}
	for k := range TestCases {
		res = append(res, k)
	}
	return res
}

func ImplementationList() []string {
	res := []string{}
	for k := range Implementations {
		res = append(res, k)
	}
	return res
}

var TestCases = map[string]TestCase{
	"VariableAvailableCapacitySingleFlow": {
		composeFileString: composeFileStringOne,
		duration:          100 * time.Second,
		leftRouter: []tcPhase{
			{
				Duration: 40 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "1000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 20 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "2500000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 20 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "600000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 20 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "1000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
		},
		rightRouter: []tcPhase{},
		plotFunc: func(outputDir, plotDir string, basetime int64) error {
			if err := os.MkdirAll(plotDir, 0755); err != nil {
				return err
			}

			for _, plot := range []string{
				"rates",
				"gcc",
				"html",
			} {
				plotCMD := exec.Command(
					"./plot.py",
					plot,
					"--name", "forward_0",
					"--input_dir", path.Join(outputDir, "forward_0"),
					"--output_dir", plotDir,
					"--basetime", fmt.Sprintf("%v", basetime),
					"--router", path.Join(outputDir, "leftrouter.log"),
				)
				fmt.Println(plotCMD.Args)
				plotCMD.Stderr = os.Stderr
				plotCMD.Stdout = os.Stdout
				if err := plotCMD.Run(); err != nil {
					return err
				}
			}
			return nil
		},
	},
	"VariableAvailableCapacityMultipleFlow": {
		composeFileString: composeFileStringTwo,
		duration:          125 * time.Second,
		leftRouter: []tcPhase{
			{
				Duration: 25 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "4000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 25 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "2000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 25 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "3500000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 25 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "1000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 25 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "2000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
		},
		rightRouter: []tcPhase{},
		plotFunc: func(outputDir, plotDir string, basetime int64) error {
			for _, direction := range []string{
				"forward_0",
				"forward_1",
			} {
				for _, plot := range []string{
					"rates",
					"gcc",
					"html",
				} {
					if err := os.MkdirAll(plotDir, 0755); err != nil {
						return err
					}
					plotCMD := exec.Command(
						"./plot.py",
						plot,
						"--name", direction,
						"--input_dir", path.Join(outputDir, direction),
						"--output_dir", plotDir,
						"--basetime", fmt.Sprintf("%v", basetime),
						"--router", path.Join(outputDir, "leftrouter.log"),
					)
					fmt.Println(plotCMD.Args)
					plotCMD.Stderr = os.Stderr
					plotCMD.Stdout = os.Stdout
					if err := plotCMD.Run(); err != nil {
						return err
					}
				}
			}
			return nil
		},
	},
	"CongestedFeedbackLinkWithBiDirectionalMediaFlows": {
		composeFileString: composeFileStringThree,
		duration:          100 * time.Second,
		leftRouter: []tcPhase{
			{
				Duration: 20 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "2000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 20 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "1000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 20 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "500000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 40 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "2000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
		},
		rightRouter: []tcPhase{
			{
				Duration: 35 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "2000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 35 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "800000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 30 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "2000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
		},
		plotFunc: func(outputDir string, plotDir string, basetime int64) error {
			for direction, router := range map[string]string{
				"forward_0":  "leftrouter.log",
				"backward_0": "rightrouter.log",
			} {
				for _, plot := range []string{
					"rates",
					"gcc",
					"html",
				} {
					if err := os.MkdirAll(plotDir, 0755); err != nil {
						return err
					}
					plotCMD := exec.Command(
						"./plot.py",
						plot,
						"--name", direction,
						"--input_dir", path.Join(outputDir, direction),
						"--output_dir", plotDir,
						"--basetime", fmt.Sprintf("%v", basetime),
						"--router", path.Join(outputDir, router),
					)
					fmt.Println(plotCMD.Args)
					plotCMD.Stderr = os.Stderr
					plotCMD.Stdout = os.Stdout
					if err := plotCMD.Run(); err != nil {
						return err
					}
				}
			}
			return nil
		},
	},
	"MediaFlowCompetingWithALongTCPFlow": {
		composeFileString: composeFileStringSix,
		duration:          120 * time.Second,
		leftRouter: []tcPhase{
			{
				Duration: 120 * time.Second,
				Config: tcConfig{
					Delay:   50 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "2000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
		},
		rightRouter: []tcPhase{},
		plotFunc: func(outputDir, plotDir string, basetime int64) error {
			if err := os.MkdirAll(plotDir, 0755); err != nil {
				return err
			}

			for plot, direction := range map[string]string{
				"rates": "forward_0",
				"gcc":   "forward_0",
				"tcp":   "forward_1",
			} {
				plotCMD := exec.Command(
					"./plot.py",
					plot,
					"--name", direction,
					"--input_dir", path.Join(outputDir, direction),
					"--output_dir", plotDir,
					"--basetime", fmt.Sprintf("%v", basetime),
					"--router", path.Join(outputDir, "leftrouter.log"),
				)
				fmt.Println(plotCMD.Args)
				plotCMD.Stderr = os.Stderr
				plotCMD.Stdout = os.Stdout
				if err := plotCMD.Run(); err != nil {
					return err
				}
			}
			plotCMD := exec.Command(
				"./plot.py",
				"html",
				"--output_dir", plotDir,
			)
			fmt.Println(plotCMD.Args)
			plotCMD.Stderr = os.Stderr
			plotCMD.Stdout = os.Stdout
			if err := plotCMD.Run(); err != nil {
				return err
			}
			return nil
		},
	},
}

type Implementation struct {
	sender       string
	senderArgs   string
	receiver     string
	receiverArgs string
}

var Implementations = map[string]Implementation{
	"pion-gcc": {
		sender:       "engelbart/bwe-test-pion",
		senderArgs:   "",
		receiver:     "engelbart/bwe-test-pion",
		receiverArgs: "",
	},
}

func (tc *TestCase) Run(ctx context.Context, implementationName string, outputDir string) error {
	implementation, ok := Implementations[implementationName]
	if !ok {
		return fmt.Errorf("unknown implementation: %v", implementationName)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	composeFile, err := os.Create("docker-compose.yml")
	if err != nil {
		return err
	}
	if _, err = composeFile.WriteString(tc.composeFileString); err != nil {
		return err
	}
	if err = composeFile.Sync(); err != nil {
		return err
	}
	defer os.Remove(composeFile.Name())

	leftRouterLog, err := os.Create(path.Join(outputDir, "leftrouter.log"))
	if err != nil {
		return err
	}
	rightRouterLog, err := os.Create(path.Join(outputDir, "rightrouter.log"))
	if err != nil {
		return err
	}

	if err = createNetwork(ctx, composeFile.Name(), tc.leftRouter, tc.rightRouter, leftRouterLog, rightRouterLog); err != nil {
		return err
	}

	cmd := exec.Command(
		"docker-compose", "-f", composeFile.Name(), "up", //"--abort-on-container-exit",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()

	for k, v := range map[string]string{
		"SENDER_0":        implementation.sender,
		"SENDER_0_ARGS":   implementation.senderArgs,
		"RECEIVER_0":      implementation.receiver,
		"RECEIVER_0_ARGS": implementation.receiverArgs,
		"SENDER_1":        implementation.sender,
		"SENDER_1_ARGS":   implementation.senderArgs,
		"RECEIVER_1":      implementation.receiver,
		"RECEIVER_1_ARGS": implementation.receiverArgs,
		"OUTPUT":          outputDir,
	} {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%v=%v", k, v))
	}

	fmt.Println(cmd.Args)
	if err = cmd.Start(); err != nil {
		return err
	}

	errCh := make(chan error)
	go func() {
		errCh <- cmd.Wait()
	}()
	select {
	case <-time.After(tc.duration + 10*time.Second):
	case <-ctx.Done():
	case err = <-errCh:
		if err != nil {
			return err
		}
	}

	return teardown(composeFile.Name())
}

func (tc *TestCase) Plot(outputDir, plotDir string, basetime int64) error {
	return tc.plotFunc(outputDir, plotDir, basetime)
}

func createNetwork(
	ctx context.Context,
	composeFile string,
	leftPhases []tcPhase,
	rightPhases []tcPhase,
	leftRouterLog io.Writer,
	rightRouterLog io.Writer,
) error {
	cmd := exec.Command(
		"docker-compose", "-f", composeFile, "up", "--force-recreate", "leftrouter", "rightrouter",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println(cmd.Args)
	if err := cmd.Start(); err != nil {
		return err
	}

	lrShaper, err := newTrafficShaper(ctx, "/leftrouter", leftPhases, leftRouterLog)
	if err != nil {
		return err
	}

	rrShaper, err := newTrafficShaper(ctx, "/rightrouter", rightPhases, rightRouterLog)
	if err != nil {
		return err
	}

	go func() {
		if err := lrShaper.run(ctx); err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		if err := rrShaper.run(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	return nil
}

func teardown(composeFile string) error {
	cmd := exec.Command("docker-compose", "-f", composeFile, "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Use host env
	cmd.Env = os.Environ()
	fmt.Println()
	fmt.Println(cmd.Args)
	fmt.Println()
	if err := cmd.Run(); err != nil {
		log.Printf("failed to shutdown docker compose setup: %v\n", err)
	}
	return nil
}
