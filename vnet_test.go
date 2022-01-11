package main

import (
	"context"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/mengelbart/bwe-test-runner/vnet"
	"github.com/pion/transport/test"
	"github.com/stretchr/testify/assert"
)

func TestVnet(t *testing.T) {
	iName := "pion-gcc-vnet"

	date := time.Now().Unix()

	for tcName, tc := range vnet.TestCases {
		t.Run(tcName, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			report := test.CheckRoutines(t)
			defer report()

			basetime := time.Now().Unix()
			outputDir := path.Join("output/", fmt.Sprintf("%v", date), iName, tcName)
			plotDir := path.Join("html/", fmt.Sprintf("%v", date), iName, tcName)

			assert.NoError(t, tc.Run(ctx, iName, outputDir))
			assert.NoError(t, tc.Plot(outputDir, plotDir, basetime))
		})
	}
}
