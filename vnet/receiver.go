package vnet

import (
	"fmt"
	"io"
	"time"

	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/packetdump"
	"github.com/pion/interceptor/pkg/twcc"
	"github.com/pion/sdp/v2"
	"github.com/pion/transport/vnet"
	"github.com/pion/webrtc/v3"
)

type receiver struct {
	privateIP string
	publicIP  string
}

func (r *receiver) createPeer(router *vnet.Router, rtpWriter, rtcpWriter io.Writer) (*webrtc.PeerConnection, error) {
	receiveNet := vnet.NewNet(&vnet.NetConfig{
		StaticIPs: []string{r.privateIP},
		StaticIP:  "",
	})
	if err := router.AddNet(receiveNet); err != nil {
		return nil, fmt.Errorf("failed to add receiveNet to routerB: %w", err)
	}

	answerSettingEngine := webrtc.SettingEngine{}
	answerSettingEngine.SetVNet(receiveNet)
	answerSettingEngine.SetICETimeouts(time.Second, time.Second, 200*time.Millisecond)
	answerSettingEngine.SetNAT1To1IPs([]string{r.publicIP}, webrtc.ICECandidateTypeHost)

	answerMediaEngine := &webrtc.MediaEngine{}
	if err := answerMediaEngine.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}

	answerRTPDumperInterceptor, err := packetdump.NewReceiverInterceptor(
		packetdump.RTPFormatter(rtpFormat),
		packetdump.RTPWriter(rtpWriter),
	)
	if err != nil {
		return nil, err
	}
	answerRTCPDumperInterceptor, err := packetdump.NewSenderInterceptor(
		packetdump.RTCPFormatter(rtcpFormat),
		packetdump.RTCPWriter(rtcpWriter),
	)
	if err != nil {
		return nil, err
	}

	answerInterceptorRegistry := &interceptor.Registry{}
	answerInterceptorRegistry.Add(answerRTPDumperInterceptor)
	answerInterceptorRegistry.Add(answerRTCPDumperInterceptor)

	answerMediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: webrtc.TypeRTCPFBTransportCC}, webrtc.RTPCodecTypeVideo)
	if err = answerMediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: sdp.TransportCCURI}, webrtc.RTPCodecTypeVideo); err != nil {
		return nil, err
	}

	answerMediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: webrtc.TypeRTCPFBTransportCC}, webrtc.RTPCodecTypeAudio)
	if err = answerMediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: sdp.TransportCCURI}, webrtc.RTPCodecTypeAudio); err != nil {
		return nil, err
	}

	generator, err := twcc.NewSenderInterceptor(twcc.SendInterval(20 * time.Millisecond))
	if err != nil {
		return nil, err
	}

	answerInterceptorRegistry.Add(generator)
	if err != nil {
		return nil, err
	}

	answerPeerConnection, err := webrtc.NewAPI(
		webrtc.WithSettingEngine(answerSettingEngine),
		webrtc.WithMediaEngine(answerMediaEngine),
		webrtc.WithInterceptorRegistry(answerInterceptorRegistry),
	).NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, err
	}

	return answerPeerConnection, nil
}
