package docker

import (
	"log"
	"os"
	"os/exec"
	"time"
)

type tcPhase struct {
	Duration time.Duration `json:"duration"`
	Config   tcConfig      `json:"config"`
}

type tcConfig struct {
	Delay   time.Duration `json:"delay"`
	Jitter  time.Duration `json:"jitter"`
	Rate    string        `json:"rate"`
	Burst   string        `json:"burst"`
	Latency time.Duration `json:"latency"`
}

func (t tcConfig) apply(container, iface string, isFirst bool) error {
	cmd := "change"
	if isFirst {
		cmd = "add"
	}
	netemCMD := exec.Command(
		"docker",
		"exec",
		container,

		"tc", "qdisc", cmd,
		"dev", iface,
		"root", "handle", "1:",
		"netem", "delay", t.Delay.String(), // t.Jitter.String(), "distribution", "normal",
	)

	log.Printf("applying tc netem: %v\n", netemCMD.Args)
	netemCMD.Stdout = os.Stdout
	netemCMD.Stderr = os.Stderr

	err := netemCMD.Run()
	if err != nil {
		return err
	}

	tbfCMD := exec.Command(
		"docker",
		"exec",
		container,

		"tc", "qdisc", cmd,
		"dev", iface,
		"parent", "1:", "handle", "2:",
		"tbf", "rate", t.Rate, "burst", t.Burst, "latency", t.Latency.String(),
	)

	log.Printf("applying tc tbf config: %v\n", tbfCMD.Args)
	tbfCMD.Stdout = os.Stdout
	tbfCMD.Stderr = os.Stderr

	err = tbfCMD.Run()
	if err != nil {
		return err
	}

	return nil
}
