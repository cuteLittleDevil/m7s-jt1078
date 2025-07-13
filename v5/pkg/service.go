package pkg

import (
	"context"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"log/slog"
	"m7s.live/v5"
	"net"
	"time"
)

func NewService(addr string, log *slog.Logger, opts ...Option) *Service {
	options := &Options{
		pubFunc: func(ctx context.Context, pack *jt1078.Packet) (publisher *m7s.Publisher, err error) {
			return nil, nil
		},
	}
	for _, op := range opts {
		op.F(options)
	}
	s := &Service{
		Logger: log,
		addr:   addr,
		opts:   options,
	}
	return s
}

type Service struct {
	*slog.Logger
	addr string
	opts *Options
}

func (s *Service) Run() {
	listen, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.Error("listen error",
			slog.String("addr", s.addr),
			slog.String("err", err.Error()))
		return
	}
	s.Info("listen tcp",
		slog.String("addr", s.addr),
		slog.String("join", s.opts.onJoinURL),
		slog.String("leave", s.opts.onLeaveURL))
	for {
		conn, err := listen.Accept()
		if err != nil {
			s.Warn("accept error",
				slog.String("err", err.Error()))
			return
		}
		client := newConnection(conn, s.Logger, s.opts.ptsFunc)
		var (
			httpBody = map[string]any{}
		)
		ctx, cancel := context.WithCancel(context.Background())
		client.onJoinEvent = func(c *connection, pack *jt1078.Packet) error {
			publisher, err := s.opts.pubFunc(ctx, pack)
			if err != nil {
				return err
			}
			c.publisher = publisher
			httpBody = map[string]any{
				"streamPath": c.publisher.StreamPath,
				"sim":        pack.Sim,
				"channel":    pack.LogicChannel,
				"startTime":  time.Now().Format(time.DateTime),
			}
			onNoticeEvent(s.opts.onJoinURL, httpBody)
			return nil
		}
		client.onLeaveEvent = func() {
			if len(httpBody) > 0 {
				httpBody["endTime"] = time.Now().Format(time.DateTime)
				onNoticeEvent(s.opts.onLeaveURL, httpBody)
			}
			cancel()
		}
		go func(ctx context.Context, waitSubscriberOverTime time.Duration) {
			if err := client.run(ctx, waitSubscriberOverTime); err != nil {
				s.Warn("run error",
					slog.Any("http body", httpBody),
					slog.String("err", err.Error()))
			}
		}(ctx, s.opts.overTime)
	}
}
