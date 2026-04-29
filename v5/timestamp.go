package v5

import (
	"time"

	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
)

// newPacketTimestamp returns a timestamp function that uses the JT1078 packet's
// own Timestamp field (uint64 milliseconds), normalized relative to the first
// packet. This works for both realtime and playback streams.
func newPacketTimestamp() func(pack *jt1078.Packet) time.Duration {
	var baseTS uint64
	var set bool
	return func(pack *jt1078.Packet) time.Duration {
		if !set {
			baseTS = pack.Timestamp
			set = true
		}
		return time.Duration(pack.Timestamp-baseTS) * time.Millisecond
	}
}
