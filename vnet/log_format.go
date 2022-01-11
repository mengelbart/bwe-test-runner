package vnet

import (
	"fmt"
	"time"

	"github.com/pion/interceptor"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
)

func rtpFormat(pkt *rtp.Packet, attributes interceptor.Attributes) string {
	// TODO(mathis): Replace timestamp by attributes.GetTimestamp as soon as
	// implemented in interceptors
	return fmt.Sprintf("%v, %v, %v, %v, %v, %v, %v\n",
		time.Now().UnixMilli(),
		pkt.PayloadType,
		pkt.SSRC,
		pkt.SequenceNumber,
		pkt.Timestamp,
		pkt.Marker,
		pkt.MarshalSize(),
	)
}

func rtcpFormat(pkts []rtcp.Packet, attributes interceptor.Attributes) string {
	// TODO(mathis): Replace timestamp by attributes.GetTimestamp as soon as
	// implemented in interceptors
	res := fmt.Sprintf("%v\t", time.Now().UnixMilli())
	for _, pkt := range pkts {
		switch feedback := pkt.(type) {
		case *rtcp.TransportLayerCC:
			res += feedback.String()
		}
	}
	return res
}
