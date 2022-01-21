package main

import (
	"context"
	"fmt"
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
		for tcName, tc := range docker.TestCases {
			for name := range docker.Implementations {
				t.Run(fmt.Sprintf("%v/%v/%v", tcName, name, video), func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					basetime := time.Now().Unix()

					outputDir := path.Join("output/", fmt.Sprintf("%v", date), tcName, name, video)
					plotDir := path.Join("html/", fmt.Sprintf("%v", date), tcName, name, video)

					assert.NoError(t, tc.Run(ctx, name, video, outputDir))
					assert.NoError(t, tc.Plot(video, outputDir, plotDir, basetime))

					assert.NoError(t, os.RemoveAll(path.Join(outputDir, "forward_0", "sink")))
					assert.NoError(t, os.RemoveAll(path.Join(outputDir, "forward_1", "sink")))
					assert.NoError(t, os.RemoveAll(path.Join(outputDir, "backward_0", "sink")))
					assert.NoError(t, os.RemoveAll(path.Join(outputDir, "backward_1", "sink")))
				})
			}
		}
	}
}
