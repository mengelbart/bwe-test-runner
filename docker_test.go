package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/mengelbart/bwe-test-runner/docker"
	"github.com/stretchr/testify/assert"
)

func TestOne(t *testing.T) {

	date := time.Now().Unix()

	for _, video := range []string{
		"foreman_concat.y4m",
		"FourPeople_concat.y4m",
		"sintel.y4m",
	} {
		//for tcName, tc := range docker.TestCases {
		for tcName := range docker.TestCases {
			for name := range docker.Implementations {
				t.Run(fmt.Sprintf("%v/%v/%v", tcName, name, video), func(t *testing.T) {
					//ctx, cancel := context.WithCancel(context.Background())
					//defer cancel()

					//basetime := time.Now().Unix()

					outputDir := path.Join("output/", fmt.Sprintf("%v", date), tcName, name, video)
					//plotDir := path.Join("html/", fmt.Sprintf("%v", date), tcName, name, video)

					//assert.NoError(t, tc.Run(ctx, name, video, outputDir))
					//assert.NoError(t, tc.Plot(video, outputDir, plotDir, basetime))

					assert.NoError(t, os.RemoveAll(path.Join(outputDir, "forward_0", "sink")))
					assert.NoError(t, os.RemoveAll(path.Join(outputDir, "forward_1", "sink")))
					assert.NoError(t, os.RemoveAll(path.Join(outputDir, "backward_0", "sink")))
					assert.NoError(t, os.RemoveAll(path.Join(outputDir, "backward_1", "sink")))
				})
			}
		}
	}
}

func TestList(t *testing.T) {
	date := time.Now().Unix()
	tests := []struct {
		name           string
		implementation string
		video          string
	}{
		{
			name:           "VariableAvailableCapacitySingleFlow50msOWD",
			implementation: "rtp-over-quic-udp-scream",
			video:          "FourPeople_concat.y4m",
		},
		{
			name:           "VariableAvailableCapacitySingleFlow50msOWD",
			implementation: "rtp-over-quic-udp-scream",
			video:          "foreman_concat.y4m",
		},
		{
			name:           "VariableAvailableCapacitySingleFlow50msOWD",
			implementation: "rtp-over-quic-udp-scream",
			video:          "sintel.y4m",
		},
		{
			name:           "VariableAvailableCapacitySingleFlow50msOWD",
			implementation: "rtp-over-quic-udp-gcc",
			video:          "sintel.y4m",
		},
		{
			name:           "VariableAvailableCapacitySingleFlow50msOWD",
			implementation: "rtp-over-quic-scream",
			video:          "sintel.y4m",
		},
		{
			name:           "VariableAvailableCapacitySingleFlow50msOWD",
			implementation: "rtp-over-quic-scream-local-feedback",
			video:          "sintel.y4m",
		},
		{
			name:           "VariableAvailableCapacitySingleFlow300msOWD",
			implementation: "rtp-over-quic-scream",
			video:          "sintel.y4m",
		},
		{
			name:           "VariableAvailableCapacitySingleFlow300msOWD",
			implementation: "rtp-over-quic-scream-local-feedback",
			video:          "sintel.y4m",
		},
		{
			name:           "VariableAvailableCapacitySingleFlow50msOWD",
			implementation: "rtp-over-quic-scream-newreno",
			video:          "sintel.y4m",
		},
		{
			name:           "VariableAvailableCapacitySingleFlow50msOWD",
			implementation: "rtp-over-quic-scream-newreno-stream",
			video:          "sintel.y4m",
		},
		{
			name:           "MediaFlowCompetingWithALongTCPRenoFlow",
			implementation: "rtp-over-quic-scream-newreno",
			video:          "sintel.y4m",
		},
		{
			name:           "MediaFlowCompetingWithALongTCPCubicFlow",
			implementation: "rtp-over-quic-scream-newreno",
			video:          "sintel.y4m",
		},
		{
			name:           "MediaFlowCompetingWithALongTCPBBRFlow",
			implementation: "rtp-over-quic-scream-newreno",
			video:          "sintel.y4m",
		},
		{
			name:           "VariableAvailableCapacitySingleFlow50msOWD",
			implementation: "rtp-over-quic-prio-scream-newreno-stream",
			video:          "sintel.y4m",
		},
		{
			name:           "VariableAvailableCapacitySingleFlow150msOWD",
			implementation: "rtp-over-quic-prio-scream-newreno-stream",
			video:          "sintel.y4m",
		},
		{
			name:           "VariableAvailableCapacitySingleFlow50msOWD",
			implementation: "rtp-over-quic-gcc-newreno-stream",
			video:          "sintel.y4m",
		},
	}
	for i, config := range tests {
		t.Run(fmt.Sprintf("%v/%v/%v", config.name, config.implementation, config.video), func(t *testing.T) {
			fmt.Printf("Test %v/%v\n", i, len(tests))
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			basetime := time.Now().Unix()

			outputDir := path.Join("output/", fmt.Sprintf("%v", date), config.name, config.implementation, config.video)
			plotDir := path.Join("html/", fmt.Sprintf("%v", date), config.name, config.implementation, config.video)

			tc := docker.TestCases[config.name]

			assert.NoError(t, tc.Run(ctx, config.implementation, config.video, outputDir))

			configJSON, err := json.MarshalIndent(struct {
				Name           string
				Implementation string
				Video          string
				Basetime       int64
			}{
				Name:           config.name,
				Implementation: config.implementation,
				Video:          config.video,
				Basetime:       basetime,
			}, "", "  ")
			assert.NoError(t, err)
			err = ioutil.WriteFile(path.Join(outputDir, "config.json"), configJSON, 0644)
			assert.NoError(t, err)

			assert.NoError(t, tc.Plot(config.video, outputDir, plotDir, basetime))

			assert.NoError(t, os.RemoveAll(path.Join(outputDir, "forward_0", "sink")))
			assert.NoError(t, os.RemoveAll(path.Join(outputDir, "forward_1", "sink")))
			assert.NoError(t, os.RemoveAll(path.Join(outputDir, "backward_0", "sink")))
			assert.NoError(t, os.RemoveAll(path.Join(outputDir, "backward_1", "sink")))
		})
	}
}
