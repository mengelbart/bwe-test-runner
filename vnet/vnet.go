package vnet

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"

	"github.com/mengelbart/syncodec"
	"github.com/pion/logging"
	"github.com/pion/transport/vnet"
	"github.com/pion/webrtc/v3"
)

const (
	defaultReferenceCapacity = 1 * vnet.MBit
	defaultMaxBurst          = 100 * vnet.KBit

	leftCIDR = "10.0.1.0/24"

	leftPublicIP1  = "10.0.1.1"
	leftPrivateIP1 = "10.0.1.101"

	leftPublicIP2  = "10.0.1.2"
	leftPrivateIP2 = "10.0.1.102"

	leftPublicIP3  = "10.0.1.3"
	leftPrivateIP3 = "10.0.1.103"

	rightCIDR = "10.0.2.0/24"

	rightPublicIP1  = "10.0.2.1"
	rightPrivateIP1 = "10.0.2.101"

	rightPublicIP2  = "10.0.2.2"
	rightPrivateIP2 = "10.0.2.102"

	rightPublicIP3  = "10.0.2.3"
	rightPrivateIP3 = "10.0.2.103"
)

var (
	defaultVideotrack = trackConfig{
		capability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8},
		id:         "video1",
		streamID:   "pion",
		vbr:        true,
		codec: func(w syncodec.FrameWriter, opts ...syncodec.StatisticalCodecOption) (syncodec.Codec, error) {
			opts = append(opts, syncodec.WithInitialTargetBitrate(1*vnet.MBit))
			return syncodec.NewStatisticalEncoder(w, opts...)
		},
		startAfter: 0,
	}
	defaultAudioTrack = trackConfig{
		capability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		id:         "audio1",
		streamID:   "pion",
		codec: func(w syncodec.FrameWriter, opts ...syncodec.StatisticalCodecOption) (syncodec.Codec, error) {
			opts = append(opts, syncodec.WithInitialTargetBitrate(20*vnet.KBit))
			return syncodec.NewStatisticalEncoder(w, opts...)
		},
		vbr:        false,
		startAfter: 0,
	}
)

var Implementations = map[string]struct{}{}

type bandwidthVariationPhase struct {
	duration      time.Duration
	capacityRatio float64
}

type trackConfig struct {
	capability webrtc.RTPCodecCapability
	id         string
	streamID   string
	codec      func(syncodec.FrameWriter, ...syncodec.StatisticalCodecOption) (syncodec.Codec, error)
	vbr        bool
	startAfter time.Duration
}

type senderReceiverPair struct {
	sender   sender
	receiver receiver
}

type Testcase struct {
	referenceCapacity int
	totalDuration     time.Duration
	left              routerConfig
	right             routerConfig
	forward           []senderReceiverPair
	backward          []senderReceiverPair
	forwardPhases     []bandwidthVariationPhase
	backwardPhases    []bandwidthVariationPhase
}

func getRTPLogWriters(rtpFilename, rtcpFilename string) (rtp, rtcp io.WriteCloser, err error) {
	rtp, err = os.Create(rtpFilename)
	if err != nil {
		return nil, nil, err
	}
	rtcp, err = os.Create(rtcpFilename)
	if err != nil {
		return nil, nil, err
	}
	return
}

func getOnTrackFunc(log logging.LeveledLogger, receivedMetrics chan<- int) func(trackRemote *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) {
	return func(trackRemote *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) {
		for {
			if err := rtpReceiver.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
				log.Errorf("failed to SetReadDeadline for rtpReceiver: %v", err)
			}
			if err := trackRemote.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
				log.Errorf("failed to SetReadDeadline for trackRemote: %v", err)
			}

			p, _, err := trackRemote.ReadRTP()
			if err == io.EOF {
				log.Info("trackRemote.ReadRTP received EOF")
				return
			}
			if err != nil {
				log.Infof("trackRemote.ReadRTP returned error: %v", err)
				return
			}
			receivedMetrics <- p.MarshalSize()
		}
	}
}

func (tc *Testcase) Plot(outputDir, plotDir string, basetime int64) error {

	for i := range tc.forward {
		logDir := path.Join(outputDir, fmt.Sprintf("forward_%v", i))
		out := path.Join(plotDir, fmt.Sprintf("forward_%v", i))
		if err := os.MkdirAll(out, 0755); err != nil {
			return err
		}
		for _, plot := range []string{
			"rates",
		} {
			plotCMD := exec.Command(
				"./plot.py",
				plot,
				"--input_dir", logDir,
				"--output_dir", out,
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
	for i := range tc.backward {
		logDir := path.Join(outputDir, fmt.Sprintf("backward_%v", i))
		out := path.Join(plotDir, fmt.Sprintf("backward_%v", i))
		for _, plot := range []string{
			"rates",
		} {
			plotCMD := exec.Command(
				"./plot.py",
				plot,
				"--input_dir", logDir,
				"--output_dir", out,
				"--basetime", fmt.Sprintf("%v", basetime),
				"--router", "rightrouter.log",
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
}

func (tc *Testcase) Run(ctx context.Context, implementationName string, outputDir string) error {
	log := logging.NewDefaultLoggerFactory().NewLogger("test")
	log.Infof("starting vnet test runner, implementation: %v", implementationName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	leftRouter, rightRouter, wan, err := createNetwork(ctx, tc.left, tc.right)
	if err != nil {
		return err
	}

	leftRouter.tbf.Set(vnet.TBFRate(tc.referenceCapacity), vnet.TBFMaxBurst(defaultMaxBurst))
	rightRouter.tbf.Set(vnet.TBFRate(tc.referenceCapacity), vnet.TBFMaxBurst(defaultMaxBurst))

	receivedMetrics := make(chan int)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		bytesReceived := 0
		for {
			select {
			case <-ctx.Done():
				return
			case b := <-receivedMetrics:
				bytesReceived += b
			case <-ticker.C:
				log.Tracef("received %v bit/s", bytesReceived*8)
				bytesReceived = 0
			}
		}
	}()

	mss := []*mediaSender{}

	for i, forward := range tc.forward {

		logDir := path.Join(outputDir, fmt.Sprintf("forward_%v", i), "send_log")
		if err = os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		rtpLogFile := path.Join(logDir, "rtp_out.log")
		rtcpLogFile := path.Join(logDir, "rtcp_in.log")
		var rtpLog, rtcpLog io.WriteCloser
		rtpLog, rtcpLog, err = getRTPLogWriters(
			rtpLogFile,
			rtcpLogFile,
		)
		if err != nil {
			return err
		}
		defer rtpLog.Close()
		defer rtcpLog.Close()

		var ccLog io.WriteCloser
		ccLog, err = os.Create(path.Join(logDir, "cc.log"))
		if err != nil {
			return err
		}
		defer ccLog.Close()

		ms := newMediaSender(log, ccLog)
		var spc *webrtc.PeerConnection
		spc, err = forward.sender.createPeer(leftRouter.Router, ms.onNewBWE, rtpLog, rtcpLog)
		if err != nil {
			return err
		}
		ms.setPeerConnection(spc)
		for _, track := range forward.sender.tracks {
			if err = ms.addTrack(track); err != nil {
				return err
			}
		}
		mss = append(mss, ms)

		logDir = path.Join(outputDir, fmt.Sprintf("forward_%v", i), "receive_log")
		if err = os.MkdirAll(logDir, 0755); err != nil {
			return err
		}
		rtpLog, rtcpLog, err = getRTPLogWriters(
			path.Join(logDir, "rtp_in.log"),
			path.Join(logDir, "rtcp_out.log"),
		)
		if err != nil {
			return err
		}
		var rpc *webrtc.PeerConnection
		rpc, err = forward.receiver.createPeer(rightRouter.Router, rtpLog, rtcpLog)
		if err != nil {
			return err
		}

		rpc.OnTrack(getOnTrackFunc(log, receivedMetrics))

		wg := untilConnectionState(webrtc.PeerConnectionStateConnected, spc, rpc)
		if err = signalPair(spc, rpc); err != nil {
			return err
		}
		defer closePairNow(spc, rpc)
		wg.Wait()
	}

	for i, backwardPair := range tc.backward {

		logDir := path.Join(outputDir, fmt.Sprintf("backward_%v", i), "send_log")
		if err = os.MkdirAll(logDir, 0755); err != nil {
			return err
		}
		var rtpLog, rtcpLog io.WriteCloser
		rtpLog, rtcpLog, err = getRTPLogWriters(
			path.Join(logDir, "rtp_out.log"),
			path.Join(logDir, "rtcp_in.log"),
		)
		if err != nil {
			return err
		}
		defer rtpLog.Close()
		defer rtcpLog.Close()

		var ccLog io.WriteCloser
		ccLog, err = os.Create(path.Join(logDir, "cc.log"))
		if err != nil {
			return err
		}
		defer ccLog.Close()

		ms := newMediaSender(log, ccLog)
		var spc *webrtc.PeerConnection
		spc, err = backwardPair.sender.createPeer(rightRouter.Router, ms.onNewBWE, rtpLog, rtcpLog)
		if err != nil {
			return err
		}
		ms.setPeerConnection(spc)
		for _, track := range backwardPair.sender.tracks {
			if err = ms.addTrack(track); err != nil {
				return err
			}
		}
		mss = append(mss, ms)

		logDir = path.Join(outputDir, fmt.Sprintf("backward_%v", i), "receive_log")
		if err = os.MkdirAll(logDir, 0755); err != nil {
			return err
		}
		rtpLog, rtcpLog, err = getRTPLogWriters(
			path.Join(logDir, "rtp_in.log"),
			path.Join(logDir, "rtcp_out.log"),
		)
		if err != nil {
			return err
		}
		var rpc *webrtc.PeerConnection
		rpc, err = backwardPair.receiver.createPeer(leftRouter.Router, rtpLog, rtcpLog)
		if err != nil {
			return err
		}

		rpc.OnTrack(getOnTrackFunc(log, receivedMetrics))

		wg := untilConnectionState(webrtc.PeerConnectionStateConnected, spc, rpc)
		if err = signalPair(spc, rpc); err != nil {
			return err
		}
		defer closePairNow(spc, rpc)
		wg.Wait()
	}

	for _, ms := range mss {
		go ms.start()
	}

	leftRouterLog, err := os.Create(path.Join(outputDir, "leftrouter.log"))
	if err != nil {
		return err
	}
	rightRouterLog, err := os.Create(path.Join(outputDir, "rightrouter.log"))
	if err != nil {
		return err
	}
	go func() {
		var nextRate int
		for _, phase := range tc.forwardPhases {
			nextRate = int(float64(tc.referenceCapacity) * phase.capacityRatio)
			rightRouter.tbf.Set(vnet.TBFRate(nextRate), vnet.TBFMaxBurst(defaultMaxBurst))
			log.Tracef("updated forward link capacity to %v", nextRate)
			fmt.Fprintf(leftRouterLog, "%v, %v\n", time.Now().UnixMilli(), nextRate)
			select {
			case <-ctx.Done():
				fmt.Fprintf(leftRouterLog, "%v, %v\n", time.Now().UnixMilli(), nextRate)
				return
			case <-time.After(phase.duration):
			}
		}
		fmt.Fprintf(leftRouterLog, "%v, %v\n", time.Now().UnixMilli(), nextRate)
	}()
	go func() {
		var nextRate int
		for _, phase := range tc.backwardPhases {
			nextRate = int(float64(tc.referenceCapacity) * phase.capacityRatio)
			leftRouter.tbf.Set(vnet.TBFRate(nextRate), vnet.TBFMaxBurst(defaultMaxBurst))
			log.Tracef("updated backward link capacity to %v", nextRate)
			fmt.Fprintf(rightRouterLog, "%v, %v\n", time.Now().UnixMilli(), nextRate)
			select {
			case <-ctx.Done():
				fmt.Fprintf(rightRouterLog, "%v, %v\n", time.Now().UnixMilli(), nextRate)
				return
			case <-time.After(phase.duration):
			}
		}
		fmt.Fprintf(rightRouterLog, "%v, %v\n", time.Now().UnixMilli(), nextRate)
	}()

	select {
	case <-time.After(tc.totalDuration):
	case <-ctx.Done():
	}
	log.Info("stopping testcase")
	for _, ms := range mss {
		ms.stop()
	}
	return wan.Stop()
}

var TestCases = map[string]Testcase{
	"TestVariableAvailableCapacitySingleFlow": {
		referenceCapacity: defaultReferenceCapacity,
		totalDuration:     100 * time.Second,
		left: routerConfig{
			cidr:      leftCIDR,
			staticIPs: []string{fmt.Sprintf("%v/%v", leftPublicIP1, leftPrivateIP1)},
		},
		right: routerConfig{
			cidr:      rightCIDR,
			staticIPs: []string{fmt.Sprintf("%v/%v", rightPublicIP1, rightPrivateIP1)},
		},
		forward: []senderReceiverPair{
			{
				sender: sender{
					privateIP: leftPrivateIP1,
					publicIP:  leftPublicIP1,
					tracks:    []trackConfig{defaultAudioTrack, defaultVideotrack},
				},
				receiver: receiver{
					privateIP: rightPrivateIP1,
					publicIP:  rightPublicIP1,
				},
			},
		},
		backward: []senderReceiverPair{},
		forwardPhases: []bandwidthVariationPhase{
			{duration: 40 * time.Second, capacityRatio: 1},
			{duration: 20 * time.Second, capacityRatio: 2.5},
			{duration: 20 * time.Second, capacityRatio: 0.6},
			{duration: 20 * time.Second, capacityRatio: 1.0},
		},
		backwardPhases: []bandwidthVariationPhase{},
	},
	"TestVariableAvailableCapacityMultipleFlow": {
		referenceCapacity: defaultReferenceCapacity,
		totalDuration:     125 * time.Second,
		left: routerConfig{
			cidr:      leftCIDR,
			staticIPs: []string{fmt.Sprintf("%v/%v", leftPublicIP1, leftPrivateIP1), fmt.Sprintf("%v/%v", leftPublicIP2, leftPrivateIP2)},
		},
		right: routerConfig{
			cidr:      rightCIDR,
			staticIPs: []string{fmt.Sprintf("%v/%v", rightPublicIP1, rightPrivateIP1), fmt.Sprintf("%v/%v", rightPublicIP2, rightPrivateIP2)},
		},
		forward: []senderReceiverPair{
			{
				sender: sender{
					privateIP: leftPrivateIP1,
					publicIP:  leftPublicIP1,
					tracks:    []trackConfig{defaultVideotrack, defaultAudioTrack},
				},
				receiver: receiver{
					privateIP: rightPrivateIP1,
					publicIP:  rightPublicIP1,
				},
			},
			{
				sender: sender{
					privateIP: leftPrivateIP2,
					publicIP:  leftPublicIP2,
					tracks:    []trackConfig{},
				},
				receiver: receiver{
					privateIP: rightPrivateIP2,
					publicIP:  rightPublicIP2,
				},
			},
		},
		backward: []senderReceiverPair{},
		forwardPhases: []bandwidthVariationPhase{
			{duration: 25 * time.Second, capacityRatio: 2.0},
			{duration: 25 * time.Second, capacityRatio: 1.0},
			{duration: 25 * time.Second, capacityRatio: 1.75},
			{duration: 25 * time.Second, capacityRatio: 0.5},
			{duration: 25 * time.Second, capacityRatio: 1.0},
		},
		backwardPhases: []bandwidthVariationPhase{},
	},
	"TestCongestedFeedbackLinkWithBiDirectionalMediaFlows": {
		referenceCapacity: defaultReferenceCapacity,
		totalDuration:     100 * time.Second,
		left: routerConfig{
			cidr: leftCIDR,
			staticIPs: []string{
				fmt.Sprintf("%v/%v", leftPublicIP1, leftPrivateIP1),
				fmt.Sprintf("%v/%v", leftPublicIP2, leftPrivateIP2),
			},
		},
		right: routerConfig{
			cidr: rightCIDR,
			staticIPs: []string{
				fmt.Sprintf("%v/%v", rightPublicIP1, rightPrivateIP1),
				fmt.Sprintf("%v/%v", rightPublicIP2, rightPrivateIP2),
			},
		},
		forward: []senderReceiverPair{
			{
				sender: sender{
					privateIP: leftPrivateIP1,
					publicIP:  leftPublicIP1,
					tracks:    []trackConfig{defaultVideotrack, defaultAudioTrack},
				},
				receiver: receiver{
					privateIP: rightPrivateIP1,
					publicIP:  rightPublicIP1,
				},
			},
		},
		backward: []senderReceiverPair{
			{
				sender: sender{
					privateIP: rightPrivateIP2,
					publicIP:  rightPublicIP2,
					tracks:    []trackConfig{},
				},
				receiver: receiver{
					privateIP: leftPrivateIP2,
					publicIP:  leftPublicIP2,
				},
			},
		},
		forwardPhases: []bandwidthVariationPhase{
			{duration: 20 * time.Second, capacityRatio: 2.0},
			{duration: 20 * time.Second, capacityRatio: 1.0},
			{duration: 20 * time.Second, capacityRatio: 0.5},
			{duration: 40 * time.Second, capacityRatio: 2.0},
		},
		backwardPhases: []bandwidthVariationPhase{
			{duration: 35 * time.Second, capacityRatio: 2.0},
			{duration: 35 * time.Second, capacityRatio: 0.8},
			{duration: 30 * time.Second, capacityRatio: 2.0},
		},
	},
	"TestRoundTripTimeFairness": {
		referenceCapacity: 4 * vnet.MBit,
		totalDuration:     300 * time.Second,
		left: routerConfig{
			cidr: leftCIDR,
			staticIPs: []string{
				fmt.Sprintf("%v/%v", leftPublicIP1, leftPrivateIP1),
				fmt.Sprintf("%v/%v", leftPublicIP2, leftPrivateIP2),
				fmt.Sprintf("%v/%v", leftPublicIP3, leftPrivateIP3),
			},
		},
		right: routerConfig{
			cidr: rightCIDR,
			staticIPs: []string{
				fmt.Sprintf("%v/%v", rightPublicIP1, rightPrivateIP1),
				fmt.Sprintf("%v/%v", rightPublicIP2, rightPrivateIP2),
				fmt.Sprintf("%v/%v", rightPublicIP3, rightPrivateIP3),
			},
		},
		forward: []senderReceiverPair{
			{
				sender: sender{
					privateIP: leftPrivateIP1,
					publicIP:  leftPublicIP1,
					tracks:    []trackConfig{},
				},
				receiver: receiver{
					privateIP: rightPrivateIP1,
					publicIP:  rightPublicIP1,
				},
			},
			{
				sender: sender{
					privateIP: leftPrivateIP2,
					publicIP:  leftPublicIP2,
					tracks:    []trackConfig{},
				},
				receiver: receiver{
					privateIP: rightPrivateIP1,
					publicIP:  rightPublicIP1,
				},
			},
			{
				sender: sender{
					privateIP: leftPrivateIP3,
					publicIP:  leftPublicIP3,
					tracks:    []trackConfig{},
				},
				receiver: receiver{
					privateIP: rightPrivateIP3,
					publicIP:  rightPublicIP3,
				},
			},
		},
		backward:       []senderReceiverPair{},
		forwardPhases:  []bandwidthVariationPhase{},
		backwardPhases: []bandwidthVariationPhase{},
	},
}

// Below are copied from pion/webrtc

func untilConnectionState(state webrtc.PeerConnectionState, peers ...*webrtc.PeerConnection) *sync.WaitGroup {
	var triggered sync.WaitGroup
	triggered.Add(len(peers))

	hdlr := func(p webrtc.PeerConnectionState) {
		if p == state {
			triggered.Done()
		}
	}
	for _, p := range peers {
		p.OnConnectionStateChange(hdlr)
	}
	return &triggered
}

func signalPair(pcOffer *webrtc.PeerConnection, pcAnswer *webrtc.PeerConnection) error {
	return signalPairWithModification(pcOffer, pcAnswer, func(sessionDescription string) string { return sessionDescription })
}

func signalPairWithModification(pcOffer *webrtc.PeerConnection, pcAnswer *webrtc.PeerConnection, modificationFunc func(string) string) error {
	// Note(albrow): We need to create a data channel in order to trigger ICE
	// candidate gathering in the background for the JavaScript/Wasm bindings. If
	// we don't do this, the complete offer including ICE candidates will never be
	// generated.
	if _, err := pcOffer.CreateDataChannel("initial_data_channel", nil); err != nil {
		return err
	}

	offer, err := pcOffer.CreateOffer(nil)
	if err != nil {
		return err
	}
	offerGatheringComplete := webrtc.GatheringCompletePromise(pcOffer)
	if err = pcOffer.SetLocalDescription(offer); err != nil {
		return err
	}
	<-offerGatheringComplete

	offer.SDP = modificationFunc(pcOffer.LocalDescription().SDP)
	if err = pcAnswer.SetRemoteDescription(offer); err != nil {
		return err
	}

	answer, err := pcAnswer.CreateAnswer(nil)
	if err != nil {
		return err
	}
	answerGatheringComplete := webrtc.GatheringCompletePromise(pcAnswer)
	if err = pcAnswer.SetLocalDescription(answer); err != nil {
		return err
	}
	<-answerGatheringComplete
	return pcOffer.SetRemoteDescription(*pcAnswer.LocalDescription())
}

// TODO: handle errors?
func closePairNow(pc1, pc2 io.Closer) {
	if err := pc1.Close(); err != nil {
		log.Printf("Failed to close PeerConnection: %v", err)
	}
	if err := pc2.Close(); err != nil {
		log.Printf("Failed to close PeerConnection: %v", err)
	}
}
