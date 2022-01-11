package vnet

import (
	"github.com/mengelbart/syncodec"
	"github.com/pion/logging"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type track struct {
	trackConfig
	log       logging.LeveledLogger
	writer    *webrtc.TrackLocalStaticSample
	rtpSender *webrtc.RTPSender
	codec     syncodec.Codec
	metrics   chan<- int
}

func (t *track) WriteFrame(frame syncodec.Frame) {
	if err := t.writer.WriteSample(media.Sample{
		Data:     frame.Content,
		Duration: frame.Duration,
	}); err != nil {
		// TODO: handle error?
		panic(err)
	}
	t.metrics <- len(frame.Content)
}

func (t *track) start(metrics chan<- int) {
	t.metrics = metrics
	go t.codec.Start()
}

func (t *track) Close() error {
	return t.codec.Close()
}
