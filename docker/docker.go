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

const composeFileString = `
version: "3.8"

services:
  leftrouter:
    image: engelbart/router
    tty: true
    command: bash
    container_name: leftrouter
    networks:
      sharednet:
        ipv4_address: 172.25.0.2
      leftnet:
        ipv4_address: 172.26.0.2
    cap_add:
      - NET_ADMIN

  rightrouter:
    image: engelbart/router
    tty: true
    command: bash
    container_name: rightrouter
    networks:
      sharednet:
        ipv4_address: 172.25.0.3
      rightnet:
        ipv4_address: 172.27.0.2
    cap_add:
      - NET_ADMIN

  sender:
    image: $SENDER
    tty: true
    container_name: sender
    hostname: sender
    environment:
      ROLE: 'sender'
      ARGS: $SENDER_ARGS
      RECEIVER: '172.27.0.3'
    volumes:
      - ./$OUTPUT/send_log:/log
      - ./input:/input:ro
    networks:
      leftnet:
        ipv4_address: 172.26.0.3
    cap_add:
      - NET_ADMIN

  receiver:
    image: $RECEIVER
    tty: true
    container_name: receiver
    hostname: receiver
    environment:
      ROLE: 'receiver'
      ARGS: $RECEIVER_ARGS
      SENDER: '172.26.03'
    volumes:
      - ./$OUTPUT/receive_log:/log
      - ./$OUTPUT/sink:/output
    networks:
      rightnet:
        ipv4_address: 172.27.0.3
    cap_add:
      - NET_ADMIN

networks:
  sharednet:
    name: sharednet
    ipam:
      driver: default
      config:
        - subnet: 172.25.0.0/16
  leftnet:
    name: leftnet
    ipam:
      driver: default
      config:
        - subnet: 172.26.0.0/16
  rightnet:
    name: rightnet
    ipam:
      driver: default
      config:
        - subnet: 172.27.0.0/16
`

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

// TODO: Move somewhere else
var composeFile = "1-docker-compose.yml"

var leftPhases = []tcPhase{
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
	{
		Duration: 0,
		Config: tcConfig{
			Delay:   50 * time.Millisecond,
			Jitter:  30 * time.Millisecond,
			Rate:    "1000000",
			Burst:   "20kb",
			Latency: 300 * time.Millisecond,
		},
	},
}
var rightPhases = []tcPhase{
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
	{
		Duration: 0,
		Config: tcConfig{
			Delay:   50 * time.Millisecond,
			Jitter:  30 * time.Millisecond,
			Rate:    "1000000",
			Burst:   "20kb",
			Latency: 300 * time.Millisecond,
		},
	},
}

func Run(ctx context.Context, implementationName string, outputDir string) error {
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
	if _, err = composeFile.WriteString(composeFileString); err != nil {
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

	if err = createNetwork(ctx, composeFile.Name(), leftPhases, rightPhases, leftRouterLog, rightRouterLog); err != nil {
		return err
	}

	cmd := exec.Command(
		"docker-compose", "-f", composeFile.Name(), "up", "--abort-on-container-exit",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()

	for k, v := range map[string]string{
		"SENDER":        implementation.sender,
		"SENDER_ARGS":   implementation.senderArgs,
		"RECEIVER":      implementation.receiver,
		"RECEIVER_ARGS": implementation.receiverArgs,
		"OUTPUT":        outputDir,
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
	case <-time.After(120 * time.Second):
	case <-ctx.Done():
	case err = <-errCh:
		if err != nil {
			return err
		}
	}

	return teardown(composeFile.Name())
}

func Plot(outputDir, plotDir string, basetime int64) error {
	if err := os.MkdirAll(plotDir, 0755); err != nil {
		return err
	}

	for _, plot := range []string{
		"rates",
		"qlog-cwnd",
		"qlog-bytes-sent",
		"qlog-rtt",
		"scream",
		"html",
	} {
		plotCMD := exec.Command(
			"./plot.py",
			plot,
			"--input_dir", outputDir,
			"--output_dir", plotDir,
			"--basetime", fmt.Sprintf("%v", basetime),
			"--router", "leftrouter.log",
		)
		fmt.Println(plotCMD.Args)
		plotCMD.Stderr = os.Stderr
		plotCMD.Stdout = os.Stdout
		if err := plotCMD.Run(); err != nil {
			return err
		}
	}
	return nil
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

	go lrShaper.run(ctx)
	go rrShaper.run(ctx)

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
