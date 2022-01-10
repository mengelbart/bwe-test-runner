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

	for name := range docker.Implementations {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			basetime := time.Now().Unix()

			outputDir := path.Join("output/", fmt.Sprintf("%v", date), name, "1")
			plotDir := path.Join("html/", fmt.Sprintf("%v", date), name, "1")

			assert.NoError(t, docker.Run(ctx, name, outputDir))
			assert.NoError(t, docker.Plot(outputDir, plotDir, basetime))
		})
	}
}
