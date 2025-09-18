package pkg

import (
	"context"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"m7s.live/v5"
	"time"
)

type Option struct {
	F func(o *Options)
}

type Options struct {
	pubFunc        func(ctx context.Context, pack *jt1078.Packet) (publisher *m7s.Publisher, err error)
	timestampFunc  func(pack *jt1078.Packet) time.Duration
	intercom       bool
	sessions       *AudioManager
	onJoinURL      string
	onLeaveURL     string
	onAudioJoinURL string
	overTime       time.Duration
}

func WithPubFunc(pubFunc func(ctx context.Context,
	pack *jt1078.Packet) (publisher *m7s.Publisher, err error)) Option {
	return Option{F: func(o *Options) {
		o.pubFunc = pubFunc
	}}
}

func WithURL(onJoinURL, onLeaveURL string) Option {
	return Option{F: func(o *Options) {
		o.onJoinURL = onJoinURL
		o.onLeaveURL = onLeaveURL
	}}
}

func WithTimestampFunc(timestampFunc func(pack *jt1078.Packet) time.Duration) Option {
	return Option{F: func(o *Options) {
		o.timestampFunc = timestampFunc
	}}
}

func WithOverTime(duration time.Duration) Option {
	return Option{F: func(o *Options) {
		o.overTime = duration
	}}
}

func WithEnableIntercom(enable bool) Option {
	return Option{F: func(o *Options) {
		o.intercom = enable
	}}
}

func WithSessions(sessions *AudioManager) Option {
	return Option{F: func(o *Options) {
		o.sessions = sessions
	}}
}
