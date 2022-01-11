package vnet

import (
	"errors"
	"fmt"
	"io"
	"math"
	"sync"
	"time"

	"github.com/pion/interceptor/pkg/cc"
	"github.com/pion/logging"
	"github.com/pion/webrtc/v3"
)

type mediaSender struct {
	log logging.LeveledLogger

	lock      sync.Mutex
	pc        *webrtc.PeerConnection
	tracks    []*track
	bwe       cc.BandwidthEstimator
	bweLogger io.Writer
	cbrSum    int // sum of bitrates of all CBR tracks
	vbrCodecs int // count of VBR tracks

	wg   sync.WaitGroup
	done chan struct{}
}

func newMediaSender(log logging.LeveledLogger, bweLogger io.Writer) *mediaSender {
	return &mediaSender{
		log:       log,
		lock:      sync.Mutex{},
		pc:        nil,
		tracks:    []*track{},
		bwe:       nil,
		bweLogger: bweLogger,
		cbrSum:    0,
		vbrCodecs: 0,
		wg:        sync.WaitGroup{},
		done:      make(chan struct{}),
	}
}

func (s *mediaSender) setPeerConnection(pc *webrtc.PeerConnection) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pc = pc
}

func (s *mediaSender) onNewBWE(id string, estimator cc.BandwidthEstimator) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.bwe = estimator
}

func (s *mediaSender) addTrack(c trackConfig) error {
	trackLocalStaticSample, err := webrtc.NewTrackLocalStaticSample(c.capability, c.id, c.streamID)
	if err != nil {
		return err
	}
	rtpSender, err := s.pc.AddTrack(trackLocalStaticSample)
	if err != nil {
		return err
	}
	track := &track{
		trackConfig: c,
		log:         s.log,
		writer:      trackLocalStaticSample,
		rtpSender:   rtpSender,
		codec:       nil,
	}
	track.codec, err = c.codec(track)
	if err != nil {
		return err
	}
	s.tracks = append(s.tracks, track)
	return nil
}

func (s *mediaSender) stop() {
	for _, track := range s.tracks {
		track.Close()
	}
	close(s.done)
	s.wg.Wait()
}

func (s *mediaSender) start() {
	metrics := make(chan int)
	for _, t := range s.tracks {
		go func() {
			for {
				if _, _, err := t.rtpSender.ReadRTCP(); err != nil {
					if errors.Is(io.EOF, err) {
						s.log.Tracef("rtpSender.ReadRTCP got EOF")
						return
					}
					s.log.Errorf("rtpSender.ReadRTCP returned error: %v", err)
					return
				}
			}
		}()
		t.start(metrics)
		if t.vbr {
			s.vbrCodecs++
		} else {
			s.cbrSum += int(t.codec.GetTargetBitrate())
		}
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		bytesSent := 0
		s.bwe.OnTargetBitrateChange(func(estimate int) {
			now := time.Now()
			s.log.Tracef("sent %v bit/s", bytesSent*8)
			bytesSent = 0

			share := float64(int(estimate)-s.cbrSum) / float64(s.vbrCodecs)
			share = math.Max(0, share)

			stats := s.bwe.GetStats()
			s.log.Infof("got new estimate: %v, %v\n", estimate, stats)
			fmt.Fprintf(s.bweLogger, "%v, %v\n", now.UnixMilli(), estimate)
			for _, track := range s.tracks {
				if track.vbr {
					s.log.Tracef("set track %v to bitrate %v\n", track.id, share)
					track.codec.SetTargetBitrate(int(share))
				}
			}
		})
		for {
			select {
			case <-s.done:
				return

			case b := <-metrics:
				bytesSent += b
			}
		}
	}()
}
