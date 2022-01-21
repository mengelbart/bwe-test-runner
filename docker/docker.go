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
	plotFunc          func(input, outputDir, plotDir string, basetime int64) error
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

func prettyPrint(s string) string {
	switch s {
	case "forward":
		return "Forward Path"
	case "forward_0":
		return "Forward Path (0)"
	case "forward_1":
		return "Forward Path (1)"
	case "backward_0":
		return "Backward Path (0)"
	case "backward_1":
		return "Backward Path (1)"
	}
	return s
}

// TODO: Remove video metrics hack
func calculateVideoMetrics(mediaSrc, mediaDst, logDir string) error {
	ffmpeg := exec.Command(
		"ffmpeg",
		"-i",
		mediaDst,
		"-i",
		mediaSrc,
		"-lavfi",
		fmt.Sprintf("ssim=%v/ssim.log:eof_action=endall;[0:v][1:v]psnr=%v/psnr.log:eof_action=endall", logDir, logDir),
		"-f",
		"null",
		"-",
	)
	log.Println(ffmpeg.Args)
	ffmpeg.Stdout = os.Stdout
	ffmpeg.Stderr = os.Stderr
	return ffmpeg.Run()
}

var TestCases = map[string]TestCase{
	"VariableAvailableCapacitySingleFlow1msOWD": {
		composeFileString: composeFileStringOne,
		duration:          100 * time.Second,
		leftRouter: []tcPhase{
			{
				Duration: 40 * time.Second,
				Config: tcConfig{
					Delay:   1 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "1000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 20 * time.Second,
				Config: tcConfig{
					Delay:   1 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "2500000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 20 * time.Second,
				Config: tcConfig{
					Delay:   1 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "600000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 20 * time.Second,
				Config: tcConfig{
					Delay:   1 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "1000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
		},
		rightRouter: []tcPhase{},
		plotFunc: func(input, outputDir, plotDir string, basetime int64) error {
			if err := os.MkdirAll(plotDir, 0755); err != nil {
				return err
			}

			if err := calculateVideoMetrics(fmt.Sprintf("./input/%v", input), fmt.Sprintf("%v/forward_0/sink/output.y4m", outputDir), fmt.Sprintf("%v/forward_0", outputDir)); err != nil {
				return err
			}

			for _, plot := range []string{
				"rates",
				"psnr",
				"ssim",
				"qlog-cwnd",
				"qlog-bytes-sent",
				"qlog-rtt",
				"scream",
				"gcc",
				"html",
			} {
				plotCMD := exec.Command(
					"./plot.py",
					plot,
					"--name", prettyPrint("forward_0"),
					"--input_dir", path.Join(outputDir, "forward_0"),
					"--output_dir", path.Join(plotDir, "forward_0"),
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
	"VariableAvailableCapacitySingleFlow50msOWD": {
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
		plotFunc: func(input, outputDir, plotDir string, basetime int64) error {
			if err := os.MkdirAll(plotDir, 0755); err != nil {
				return err
			}

			if err := calculateVideoMetrics(fmt.Sprintf("./input/%v", input), fmt.Sprintf("%v/forward_0/sink/output.y4m", outputDir), fmt.Sprintf("%v/forward_0", outputDir)); err != nil {
				return err
			}

			for _, plot := range []string{
				"rates",
				"psnr",
				"ssim",
				"qlog-cwnd",
				"qlog-bytes-sent",
				"qlog-rtt",
				"scream",
				"gcc",
				"html",
			} {
				plotCMD := exec.Command(
					"./plot.py",
					plot,
					"--name", prettyPrint("forward_0"),
					"--input_dir", path.Join(outputDir, "forward_0"),
					"--output_dir", path.Join(plotDir, "forward_0"),
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
	"VariableAvailableCapacitySingleFlow150msOWD": {
		composeFileString: composeFileStringOne,
		duration:          100 * time.Second,
		leftRouter: []tcPhase{
			{
				Duration: 40 * time.Second,
				Config: tcConfig{
					Delay:   150 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "1000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 20 * time.Second,
				Config: tcConfig{
					Delay:   150 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "2500000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 20 * time.Second,
				Config: tcConfig{
					Delay:   150 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "600000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 20 * time.Second,
				Config: tcConfig{
					Delay:   150 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "1000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
		},
		rightRouter: []tcPhase{},
		plotFunc: func(input, outputDir, plotDir string, basetime int64) error {
			if err := os.MkdirAll(plotDir, 0755); err != nil {
				return err
			}

			if err := calculateVideoMetrics(fmt.Sprintf("./input/%v", input), fmt.Sprintf("%v/forward_0/sink/output.y4m", outputDir), fmt.Sprintf("%v/forward_0", outputDir)); err != nil {
				return err
			}

			for _, plot := range []string{
				"rates",
				"psnr",
				"ssim",
				"qlog-cwnd",
				"qlog-bytes-sent",
				"qlog-rtt",
				"scream",
				"gcc",
				"html",
			} {
				plotCMD := exec.Command(
					"./plot.py",
					plot,
					"--name", prettyPrint("forward_0"),
					"--input_dir", path.Join(outputDir, "forward_0"),
					"--output_dir", path.Join(plotDir, "forward_0"),
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
	"VariableAvailableCapacitySingleFlow300msOWD": {
		composeFileString: composeFileStringOne,
		duration:          100 * time.Second,
		leftRouter: []tcPhase{
			{
				Duration: 40 * time.Second,
				Config: tcConfig{
					Delay:   300 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "1000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 20 * time.Second,
				Config: tcConfig{
					Delay:   300 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "2500000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 20 * time.Second,
				Config: tcConfig{
					Delay:   300 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "600000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
			{
				Duration: 20 * time.Second,
				Config: tcConfig{
					Delay:   300 * time.Millisecond,
					Jitter:  30 * time.Millisecond,
					Rate:    "1000000",
					Burst:   "20kb",
					Latency: 300 * time.Millisecond,
				},
			},
		},
		rightRouter: []tcPhase{},
		plotFunc: func(input, outputDir, plotDir string, basetime int64) error {
			if err := os.MkdirAll(plotDir, 0755); err != nil {
				return err
			}

			if err := calculateVideoMetrics(fmt.Sprintf("./input/%v", input), fmt.Sprintf("%v/forward_0/sink/output.y4m", outputDir), fmt.Sprintf("%v/forward_0", outputDir)); err != nil {
				return err
			}

			for _, plot := range []string{
				"rates",
				"psnr",
				"ssim",
				"qlog-cwnd",
				"qlog-bytes-sent",
				"qlog-rtt",
				"scream",
				"gcc",
				"html",
			} {
				plotCMD := exec.Command(
					"./plot.py",
					plot,
					"--name", prettyPrint("forward_0"),
					"--input_dir", path.Join(outputDir, "forward_0"),
					"--output_dir", path.Join(plotDir, "forward_0"),
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
		plotFunc: func(input, outputDir, plotDir string, basetime int64) error {
			for _, direction := range []string{
				"forward_0",
				"forward_1",
			} {
				if err := calculateVideoMetrics(fmt.Sprintf("./input/%v", input), fmt.Sprintf("%v/%v/sink/output.y4m", outputDir, direction), fmt.Sprintf("%v/%v", outputDir, direction)); err != nil {
					return err
				}
				for _, plot := range []string{
					"rates",
					"psnr",
					"ssim",
					"qlog-cwnd",
					"qlog-bytes-sent",
					"qlog-rtt",
					"scream",
					"gcc",
					"html",
				} {
					if err := os.MkdirAll(plotDir, 0755); err != nil {
						return err
					}
					plotCMD := exec.Command(
						"./plot.py",
						plot,
						"--name", prettyPrint(direction),
						"--input_dir", path.Join(outputDir, direction),
						"--output_dir", path.Join(plotDir, direction),
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
		plotFunc: func(input, outputDir string, plotDir string, basetime int64) error {
			for direction, router := range map[string]string{
				"forward_0":  "leftrouter.log",
				"backward_0": "rightrouter.log",
			} {
				if err := calculateVideoMetrics(fmt.Sprintf("./input/%v", input), fmt.Sprintf("%v/%v/sink/output.y4m", outputDir, direction), fmt.Sprintf("%v/%v", outputDir, direction)); err != nil {
					return err
				}
				for _, plot := range []string{
					"rates",
					"psnr",
					"ssim",
					"qlog-cwnd",
					"qlog-bytes-sent",
					"qlog-rtt",
					"scream",
					"gcc",
					"html",
				} {
					if err := os.MkdirAll(plotDir, 0755); err != nil {
						return err
					}
					plotCMD := exec.Command(
						"./plot.py",
						plot,
						"--name", prettyPrint(direction),
						"--input_dir", path.Join(outputDir, direction),
						"--output_dir", path.Join(plotDir, direction),
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
		plotFunc: func(input, outputDir, plotDir string, basetime int64) error {
			if err := os.MkdirAll(plotDir, 0755); err != nil {
				return err
			}

			direction := "forward_0"
			if err := calculateVideoMetrics(fmt.Sprintf("./input/%v", input), fmt.Sprintf("%v/%v/sink/output.y4m", outputDir, direction), fmt.Sprintf("%v/%v", outputDir, direction)); err != nil {
				return err
			}

			for plot, direction := range map[string]string{
				"rates":           "forward_0",
				"psnr":            "forward_0",
				"ssim":            "forward_0",
				"qlog-cwnd":       "forward_0",
				"qlog-bytes-sent": "forward_0",
				"qlog-rtt":        "forward_0",
				"scream":          "forward_0",
				"gcc":             "forward_0",
				"tcp":             "forward_1",
			} {
				plotCMD := exec.Command(
					"./plot.py",
					plot,
					"--name", prettyPrint(direction),
					"--input_dir", path.Join(outputDir, direction),
					"--output_dir", path.Join(plotDir, direction),
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
	//"pion-gcc": {
	//	sender:       "engelbart/bwe-test-pion",
	//	senderArgs:   "",
	//	receiver:     "engelbart/bwe-test-pion",
	//	receiverArgs: "",
	//},
	"rtp-over-quic-udp-gcc": {
		sender:       "engelbart/rtp-over-quic",
		senderArgs:   "--transport udp --cc-dump /log/gcc.log --rtcp-dump /log/rtcp_in.log --rtp-dump /log/rtp_out.log --gcc --codec h264 --source /input/%v",
		receiver:     "engelbart/rtp-over-quic",
		receiverArgs: "--transport udp --rtcp-dump /log/rtcp_out.log --rtp-dump /log/rtp_in.log --twcc --codec h264 --sink /output/output.y4m",
	},
	"rtp-over-quic-udp-scream": {
		sender:       "engelbart/rtp-over-quic",
		senderArgs:   "--transport udp --cc-dump /log/scream.log --rtcp-dump /log/rtcp_in.log --rtp-dump /log/rtp_out.log --scream --codec h264 --source /input/%v",
		receiver:     "engelbart/rtp-over-quic",
		receiverArgs: "--transport udp --rtcp-dump /log/rtcp_out.log --rtp-dump /log/rtp_in.log --rfc8888 --codec h264 --sink /output/output.y4m",
	},
	"rtp-over-quic-tcp": {
		sender:       "engelbart/rtp-over-quic",
		senderArgs:   "--transport tcp --rtcp-dump /log/rtcp_in.log --rtp-dump /log/rtp_out.log --codec h264 --source /input/%v",
		receiver:     "engelbart/rtp-over-quic",
		receiverArgs: "--transport tcp --rtcp-dump /log/rtcp_out.log --rtp-dump /log/rtp_in.log --codec h264 --sink /output/output.y4m",
	},
	"rtp-over-quic-tcp-gcc": {
		sender:       "engelbart/rtp-over-quic",
		senderArgs:   "--transport tcp --cc-dump /log/gcc.log --rtcp-dump /log/rtcp_in.log --rtp-dump /log/rtp_out.log --gcc --codec h264 --source /input/%v",
		receiver:     "engelbart/rtp-over-quic",
		receiverArgs: "--transport tcp --rtcp-dump /log/rtcp_out.log --rtp-dump /log/rtp_in.log --twcc --codec h264 --sink /output/output.y4m",
	},
	"rtp-over-quic-tcp-scream": {
		sender:       "engelbart/rtp-over-quic",
		senderArgs:   "--transport tcp --cc-dump /log/scream.log --rtcp-dump /log/rtcp_in.log --rtp-dump /log/rtp_out.log --scream --codec h264 --source /input/%v",
		receiver:     "engelbart/rtp-over-quic",
		receiverArgs: "--transport tcp --rtcp-dump /log/rtcp_out.log --rtp-dump /log/rtp_in.log --rfc8888 --codec h264 --sink /output/output.y4m",
	},
	"rtp-over-quic-gcc": {
		sender:       "engelbart/rtp-over-quic",
		senderArgs:   "--cc-dump /log/gcc.log --rtcp-dump /log/rtcp_in.log --rtp-dump /log/rtp_out.log --qlog /log --gcc --codec h264 --source /input/%v",
		receiver:     "engelbart/rtp-over-quic",
		receiverArgs: "--rtcp-dump /log/rtcp_out.log --rtp-dump /log/rtp_in.log --qlog /log --twcc --codec h264 --sink /output/output.y4m",
	},
	"rtp-over-quic-scream": {
		sender:       "engelbart/rtp-over-quic",
		senderArgs:   "--cc-dump /log/scream.log --rtcp-dump /log/rtcp_in.log --rtp-dump /log/rtp_out.log --qlog /log --scream --codec h264 --source /input/%v",
		receiver:     "engelbart/rtp-over-quic",
		receiverArgs: "--rtcp-dump /log/rtcp_out.log --rtp-dump /log/rtp_in.log --qlog /log --rfc8888 --codec h264 --sink /output/output.y4m",
	},
	"rtp-over-quic-gcc-newreno": {
		sender:       "engelbart/rtp-over-quic",
		senderArgs:   "--cc-dump /log/gcc.log --rtcp-dump /log/rtcp_in.log --rtp-dump /log/rtp_out.log --qlog /log --gcc --newreno --codec h264 --source /input/%v",
		receiver:     "engelbart/rtp-over-quic",
		receiverArgs: "--rtcp-dump /log/rtcp_out.log --rtp-dump /log/rtp_in.log --qlog /log --twcc --codec h264 --sink /output/output.y4m",
	},
	"rtp-over-quic-scream-newreno": {
		sender:       "engelbart/rtp-over-quic",
		senderArgs:   "--cc-dump /log/scream.log --rtcp-dump /log/rtcp_in.log --rtp-dump /log/rtp_out.log --qlog /log --scream --newreno --codec h264 --source /input/%v",
		receiver:     "engelbart/rtp-over-quic",
		receiverArgs: "--rtcp-dump /log/rtcp_out.log --rtp-dump /log/rtp_in.log --qlog /log --rfc8888 --codec h264 --sink /output/output.y4m",
	},
	"rtp-over-quic-scream-local-feedback": {
		sender:       "engelbart/rtp-over-quic",
		senderArgs:   "--cc-dump /log/scream.log --rtcp-dump /log/rtcp_in.log --rtp-dump /log/rtp_out.log --qlog /log --scream --local-rfc8888 --codec h264 --source /input/%v",
		receiver:     "engelbart/rtp-over-quic",
		receiverArgs: "--rtcp-dump /log/rtcp_out.log --rtp-dump /log/rtp_in.log --qlog /log --codec h264 --sink /output/output.y4m",
	},
	"rtp-over-quic-scream-local-feedback-newreno": {
		sender:       "engelbart/rtp-over-quic",
		senderArgs:   "--cc-dump /log/scream.log --rtcp-dump /log/rtcp_in.log --rtp-dump /log/rtp_out.log --qlog /log --scream --newreno --local-rfc8888 --codec h264 --source /input/%v",
		receiver:     "engelbart/rtp-over-quic",
		receiverArgs: "--rtcp-dump /log/rtcp_out.log --rtp-dump /log/rtp_in.log --qlog /log --codec h264 --sink /output/output.y4m",
	},
	"rtp-over-quic-gcc-newreno-stream": {
		sender:       "engelbart/rtp-over-quic",
		senderArgs:   "--cc-dump /log/gcc.log --rtcp-dump /log/rtcp_in.log --rtp-dump /log/rtp_out.log --qlog /log --gcc --newreno --codec h264 --source /input/%v --stream",
		receiver:     "engelbart/rtp-over-quic",
		receiverArgs: "--rtcp-dump /log/rtcp_out.log --rtp-dump /log/rtp_in.log --qlog /log --twcc --codec h264 --sink /output/output.y4m",
	},
	"rtp-over-quic-scream-newreno-stream": {
		sender:       "engelbart/rtp-over-quic",
		senderArgs:   "--cc-dump /log/scream.log --rtcp-dump /log/rtcp_in.log --rtp-dump /log/rtp_out.log --qlog /log --scream --newreno --codec h264 --source /input/%v --stream",
		receiver:     "engelbart/rtp-over-quic",
		receiverArgs: "--rtcp-dump /log/rtcp_out.log --rtp-dump /log/rtp_in.log --qlog /log --rfc8888 --codec h264 --sink /output/output.y4m",
	},
}

func (tc *TestCase) Run(ctx context.Context, implementationName, input, outputDir string) error {
	implementation, ok := Implementations[implementationName]
	if !ok {
		return fmt.Errorf("unknown implementation: %v", implementationName)
	}

	for _, subdir := range []string{
		"forward_0/send_log",
		"forward_0/receive_log",
		"forward_0/sink",
		"forward_1/send_log",
		"forward_1/receive_log",
		"forward_1/sink",
		"backward_0/send_log",
		"backward_0/receive_log",
		"backward_0/sink",
		"backward_1/send_log",
		"backward_1/receive_log",
		"backward_1/sink",
	} {
		if err := os.MkdirAll(path.Join(outputDir, subdir), 0755); err != nil {
			return err
		}
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
		"SENDER_0_ARGS":   fmt.Sprintf(implementation.senderArgs, input),
		"RECEIVER_0":      implementation.receiver,
		"RECEIVER_0_ARGS": implementation.receiverArgs,
		"SENDER_1":        implementation.sender,
		"SENDER_1_ARGS":   fmt.Sprintf(implementation.senderArgs, input),
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

func (tc *TestCase) Plot(input, outputDir, plotDir string, basetime int64) error {
	return tc.plotFunc(input, outputDir, plotDir, basetime)
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
