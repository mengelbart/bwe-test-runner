package main

import (
	"context"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/mengelbart/bwe-test-runner/docker"
	"github.com/stretchr/testify/assert"
)

func TestOne(t *testing.T) {

	date := time.Now().Unix()

	for tcName, tc := range docker.TestCases {
		for name := range docker.Implementations {
			t.Run(fmt.Sprintf("%v-%v", tcName, name), func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				basetime := time.Now().Unix()

				outputDir := path.Join("output/", fmt.Sprintf("%v", date), name, tcName)
				plotDir := path.Join("html/", fmt.Sprintf("%v", date), name, tcName)

				assert.NoError(t, tc.Run(ctx, name, outputDir))
				assert.NoError(t, tc.Plot(outputDir, plotDir, basetime))
			})
		}
	}
}
