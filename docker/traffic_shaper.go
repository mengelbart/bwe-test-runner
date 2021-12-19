package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// tsTimeout is the time to wait for containers to spin up before timing out
const tsTimeout = 5 * time.Minute

var errTrafficShaperTimeout = errors.New("traffic shaper timed out while waiting for containers to spin up")

type TrafficShaper struct {
	log       io.WriteCloser
	container string
	iface     string
	phases    []tcPhase
}

func newTrafficShaper(ctx context.Context, name string, phases []tcPhase, log io.WriteCloser) (*TrafficShaper, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	for time.Since(start) < tsTimeout {
		var containers []types.Container
		if containers, err = cli.ContainerList(ctx, types.ContainerListOptions{}); err != nil {
			return nil, fmt.Errorf("failed to get container list: %w", err)
		}
		for _, c := range containers {
			for _, cname := range c.Names {
				if cname == name {
					iface, err := getInterfaceForIP(name, "172.25.0.0/24")
					if err != nil {
						return nil, fmt.Errorf("failed to get interface: %w", err)
					}
					return &TrafficShaper{
						log:       log,
						container: name,
						iface:     iface,
						phases:    phases,
					}, nil
				}
			}
		}
	}
	return nil, errTrafficShaperTimeout
}

func (s *TrafficShaper) run(ctx context.Context) error {
	log.Printf("run traffic shaper: '%v'/'%v'\n", s.container, s.iface)
	var lastRate string
	defer func() {
		now := time.Now()
		fmt.Fprintf(s.log, "%v, %v\n", now.UnixMilli(), lastRate)
		s.log.Close()
	}()
	if len(s.phases) == 0 {
		return nil
	}
	for i, p := range s.phases {
		lastRate = p.Config.Rate
		fmt.Fprintf(s.log, "%v, %v\n", time.Now().UnixMilli(), lastRate)
		err := p.Config.apply(s.container, s.iface, i == 0)
		if err != nil {
			return err
		}

		if p.Duration.Duration == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(p.Duration.Duration):
		}
	}
	return nil
}

func getInterfaceForIP(container, prefix string) (string, error) {
	b, err := exec.Command("docker", "exec", container, "ip", "addr", "show", "to", prefix).Output()
	if err != nil {
		return "", err
	}
	iface := strings.Split(string(b), ":")[1]
	iface = strings.Split(iface, "@")[0]
	iface = strings.TrimSpace(iface)
	return iface, nil
}
