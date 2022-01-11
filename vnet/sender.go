package vnet

import (
	"fmt"
	"io"
	"time"

	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/cc"
	"github.com/pion/interceptor/pkg/gcc"
	"github.com/pion/interceptor/pkg/packetdump"
	"github.com/pion/transport/vnet"
	"github.com/pion/webrtc/v3"
)

type sender struct {
	privateIP string
	publicIP  string

	tracks []trackConfig
}

func (s *sender) createPeer(router *vnet.Router, cb cc.NewPeerConnectionCallback, rtpWriter, rtcpWriter io.Writer) (*webrtc.PeerConnection, error) {
	sendNet := vnet.NewNet(&vnet.NetConfig{
		StaticIPs: []string{s.privateIP},
		StaticIP:  "",
	})

	if err := router.AddNet(sendNet); err != nil {
		return nil, fmt.Errorf("failed to add sendNet to routerA: %w", err)
	}

	offerSettingEngine := webrtc.SettingEngine{}
	offerSettingEngine.SetVNet(sendNet)
	offerSettingEngine.SetICETimeouts(time.Second, time.Second, 200*time.Millisecond)
	offerSettingEngine.SetNAT1To1IPs([]string{s.publicIP}, webrtc.ICECandidateTypeHost)

	offerMediaEngine := &webrtc.MediaEngine{}
	if err := offerMediaEngine.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}

	offerRTPDumperInterceptor, err := packetdump.NewSenderInterceptor(
		packetdump.RTPFormatter(rtpFormat),
		packetdump.RTPWriter(rtpWriter),
	)
	if err != nil {
		return nil, err
	}
	offerRTCPDumperInterceptor, err := packetdump.NewReceiverInterceptor(
		packetdump.RTCPFormatter(rtcpFormat),
		packetdump.RTCPWriter(rtcpWriter),
	)
	if err != nil {
		return nil, err
	}

	offerInterceptorRegistry := &interceptor.Registry{}
	offerInterceptorRegistry.Add(offerRTPDumperInterceptor)
	offerInterceptorRegistry.Add(offerRTCPDumperInterceptor)

	fx := func() (cc.BandwidthEstimator, error) {
		return gcc.NewSendSideBWE(gcc.SendSideBWEInitialBitrate(800_000), gcc.SendSideBWEPacer(gcc.NewLeakyBucketPacer(800_000)))
	}
	ccInterceptor, err := cc.NewInterceptor(fx)
	if err != nil {
		return nil, err
	}
	ccInterceptor.OnNewPeerConnection(cb)

	offerInterceptorRegistry.Add(ccInterceptor)

	err = webrtc.ConfigureTWCCHeaderExtensionSender(offerMediaEngine, offerInterceptorRegistry)
	if err != nil {
		return nil, err
	}

	offerPeerConnection, err := webrtc.NewAPI(
		webrtc.WithSettingEngine(offerSettingEngine),
		webrtc.WithMediaEngine(offerMediaEngine),
		webrtc.WithInterceptorRegistry(offerInterceptorRegistry),
	).NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, err
	}

	return offerPeerConnection, nil
}
